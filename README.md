# This a simplebank
very simple haha~

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
