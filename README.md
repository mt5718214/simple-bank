# This a simplebank
very simple haha~

## Docker-compose-healthcheck
### 1. use additional scripts [wait-for](https://github.com/eficode/wait-for)
```sh
# Download the wait-for file & rename to wait-for.sh
# make it executable
chmod +x wait-for.sh
```
```Dockerfile
FROM alpine:latest AS release
WORKDIR /app
...
COPY ./wait-for.sh . # copy into images
ENTRYPOINT ["./wait-for.sh", "postgres:5432", "--", "/app/start.sh"]
CMD ["/app/main"]
```

### 2. use the [healthcheck](https://docs.docker.com/compose/compose-file/compose-file-v3/#healthcheck) property
```yml
version: '3.9'

services:
  postgres:
    image: postgres:12-alpine
    environment:
      POSTGRES_USER: root
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: simple_bank
    # healthcheck configured
    healthcheck:
      test: ["CMD-SHELL", "pg_isready"]
      interval: 10s
      timeout: 5s
      retries: 5
    ports:
      - 5432:5432

  simplebank:
    depends_on:
      postgres:
        # This specifies that a dependency is expected to be “healthy”, which is defined with healthcheck, before starting a dependent service.
        condition: service_healthy
    build:
      context: ./
    environment:
      DB_SOURCE: 'postgres://root:secret@postgres:5432/simple_bank?sslmode=disable'
    ports:
      - 8080:8080
    entrypoint: ["/app/start.sh"]
    command: ["/app/main"]
```

## Git commit type
- feat: 新增/修改功能 (feature)。
- fix: 修補 bug (bug fix)。
- docs: 文件 (documentation)。
- style: 格式 (不影響程式碼運行的變動 white-space, formatting, missing semi - colons, etc)。
- refactor: 重構 (既不是新增功能，也不是修補 bug 的程式碼變動)。
- perf: 改善效能 (A code change that improves performance)。
- test: 增加測試 (when adding missing tests)。
- chore: 建構程序或輔助工具的變動 (maintain)。
- revert: 撤銷回覆先前的 commit 例如：revert: type(scope): subject (回覆版本：xxxx)。

## Linters

### install brew install golangci-lint
```
brew install golangci-lint
```

### add .golangci.yml setting
```
cat <<EOF>.golangci.yml
linters-settings:
  govet:
    vettool:
      settings:
        - all

linters:
  disable-all: true
  enable:
    - errcheck
    - gofmt
    - govet
    - revive
    - staticcheck
    - unused
EOF
```

### run lint
```
golangci-lint run
```


## pre-commit
### create pre-commit file in the .githooks directoty
```
$ mkdir .githooks
```
### create pre-commit file
```
$ touch pre-commit
```
```bash
# Enter the following statement

#!/bin/bash
echo "pre-commit"
# run lint
golangci-lint run
```

### make pre-commit file as executable
```
chmod 744 pre-commit
```

### let git know where to execute the hook file
```
$ git config core.hooksPath .githooks
```
