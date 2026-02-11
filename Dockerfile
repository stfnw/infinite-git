# Deploy this app with:

# # Initial version without source code (reproducible repo with commit hash
# # d3810f897df1a232a3e9cc3f434f9134b40160bf):
# rm -fv ./repo*
# D=$(mktemp -d)
# pushd $D
# DATE='1970-01-01T01:00:00+00:00'
# export GIT_AUTHOR_NAME=author_name
# export GIT_AUTHOR_EMAIL=author_email
# export GIT_AUTHOR_DATE=$DATE
# export GIT_COMMITTER_NAME=committer_name
# export GIT_COMMITTER_EMAIL=committer_email
# export GIT_COMMITTER_DATE=$DATE
# git init ; echo 'not yet' > README.md ; git add . ; git commit --no-gpg-sign -m init ; git show
# popd
# git -C $D rev-parse main | tr -d '\n' > ./repomainhash
# echo > ./repotext
# git -C $D gc ; cp $D/.git/objects/pack/pack-*.pack ./repopackfile

# # Final version:
# rm -fv ./repo*
# git rev-parse main | tr -d '\n' > ./repomainhash
# { LC_COLLATE=C tree -aC --gitignore ; find -type f | LC_COLLATE=C sort | grep -v -e '^\./repo.*' -e '\.git' | xargs -I{} batcat --color=always {} ; } > ./repotext
# git gc ; cp ./.git/objects/pack/pack-*.pack ./repopackfile

# podman build -t infinite-git .
# podman run -d --rm -p 8080:8080 localhost/infinite-git



FROM docker.io/golang:1.26.0-trixie AS build

WORKDIR /app

COPY go.mod *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /infinite-git



FROM gcr.io/distroless/base-debian13 AS release

WORKDIR /

COPY --from=build /infinite-git /infinite-git
COPY ./repo* /

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/infinite-git"]
