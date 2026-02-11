// SPDX-FileCopyrightText: 2026 Stefan Walter (stfnw)
//
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Git sha1 revision hash for main branch of this repository.
var REPODATA_MAINHASH []byte

// Sideband data for this repository: code itself rendered in text form; consists of output of:
var REPODATA_TEXT []byte

// Sideband data for this repository: actual git repository packfile.
var REPODATA_PACKFILE []byte

// See Dockerfile for how to generate/deploy these.

const MAX_REQUEST_SIZE = 0x10000

type MyHttpHandler func(*slog.Logger, http.ResponseWriter, *http.Request)

func setDefaultHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
	w.Header().Set("Server", "infinite-git")
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("pragma", "no-cache")
}

func handlerSnakeGet(logger *slog.Logger, w http.ResponseWriter, r *http.Request) {
	setDefaultHeaders(w)
	w.WriteHeader(200)

	w.Write(pktline([]byte("# service=git-upload-pack\n")))
	w.Write(flush())
	w.Write(pktline([]byte("version 2\n")))
	w.Write(pktline([]byte("agent=git/infinite\n")))
	w.Write(pktline([]byte("ls-refs=unborn\n")))
	w.Write(pktline([]byte("fetch=shallow wait-for-done filter\n")))
	w.Write(pktline([]byte("server-option\n")))
	w.Write(pktline([]byte("object-format=sha1\n")))
	w.Write(flush())
}

func handlerSnakePost(logger *slog.Logger, w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MAX_REQUEST_SIZE)
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Warn("Error reading request body")
		return
	}

	setDefaultHeaders(w)
	w.WriteHeader(200)

	if bytes.Contains(body, []byte("command=ls-refs")) {
		w.Write(pktline(append(REPODATA_MAINHASH, []byte(" HEAD symref-target:refs/heads/main\n")...)))
		w.Write(pktline(append(REPODATA_MAINHASH, []byte(" refs/heads/main\n")...)))
		w.Write(flush())

	} else if bytes.Contains(body, []byte("command=fetch")) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			logger.Warn("Streaming (Transfer-Encoding Chunked) not supported")
			return
		}

		w.Write(pktline([]byte("packfile\n")))

		for _, msg := range []string{
			"Enumerating objects: 0/1 ...\n",
			"Enumerating objects: 1/2 ...\n",
			"Enumerating objects: 2/3 ...\n",
			"Enumerating objects: 3/4 ...\n",
			"Enumerating objects: 4/999999999999999999999 ...\n",
			"Enumerating objects: 999999999999999999999/999999999999999999999, done.\n\n",
			"Booting up ...\n",
			"Done\n\n",
			"Hello World\n\n",
			"(Because of terminal escape codes, this works best on linux)\n\n",
		} {
			w.Write(sideband(SidebandProgress, []byte(msg)))
			flusher.Flush()
			time.Sleep(500 * time.Millisecond)
		}
		time.Sleep(1 * time.Second)

		for output := range SnakeGame() {
			w.Write(sideband(SidebandProgress, []byte(output)))
			flusher.Flush()
			time.Sleep(30 * time.Millisecond)
		}

		w.Write(sideband(SidebandProgress, []byte("... if 10 people write me that they're interested to [infinite-git at stfnw.de] I'll release the source code\n\n")))
		flusher.Flush()
		time.Sleep(500 * time.Millisecond)

		// w.Write(sideband(SidebandProgress, []byte("... Here's the source code\n\n")))
		// flusher.Flush()
		// time.Sleep(500 * time.Millisecond)

		for line := range bytes.SplitAfterSeq(REPODATA_TEXT, []byte{'\n'}) {
			w.Write(sideband(SidebandProgress, line))
			flusher.Flush()
			time.Sleep(70 * time.Millisecond)
		}

		w.Write(sideband(SidebandPackfile, REPODATA_PACKFILE))

		w.Write(sideband(SidebandProgress, []byte("Finished cloning.")))
		flusher.Flush()
		time.Sleep(500 * time.Millisecond)

		w.Write(flush())

	} else {
		logger.Warn("Unexpected POST request")

	}
}

// Wrap a http handler function to also log.
func logHF(logger *slog.Logger, handler MyHttpHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(logger, w, r)
	}
}

// Wrap a http handler to also log.
func logH(logger *slog.Logger, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("http request", "method", r.Method, "url", r.URL.Path, "from", r.RemoteAddr, "user-agent", r.UserAgent())
		handler.ServeHTTP(w, r)
	})
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	var err error

	REPODATA_MAINHASH, err = os.ReadFile("./repomainhash")
	if err != nil {
		logger.Error("Error reading file ./repomainhash")
		return
	}

	REPODATA_TEXT, err = os.ReadFile("./repotext")
	if err != nil {
		logger.Error("Error reading file ./repotext")
		return
	}

	REPODATA_PACKFILE, err = os.ReadFile("./repopackfile")
	if err != nil {
		logger.Error("Error reading file ./repopackfile")
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /info/refs", logHF(logger, handlerSnakeGet))
	mux.HandleFunc("POST /git-upload-pack", logHF(logger, handlerSnakePost))

	srv := http.Server{
		Addr:         ":8080",
		Handler:      logH(logger, mux),
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Minute,
	}

	// Serve only HTTP/1 (shouldn't matter but better safe than sorry
	// regarding streaming / chunking).
	srv.Protocols = new(http.Protocols)
	srv.Protocols.SetHTTP1(true)

	logger.Info("Server is running ...")

	srv.ListenAndServe()
}
