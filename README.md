# Yunxiao CLI

Yunxiao CLI 是一个面向 AI Agent 和自动化脚本的阿里云云效（Yunxiao）DevOps 命令行工具。它将云效常用只读能力封装为独立的 Go 二进制程序，便于在 shell、CI、subprocess 和其他 Agent 工作流中稳定调用。

项目重点不是提供交互式人类界面，而是提供稳定、可解析、可自动化的 CLI 契约：成功和失败都输出 JSON envelope，exit code 有明确语义，stderr 只承载诊断信息。

## 功能特性

- 独立 Go + Cobra CLI，不依赖 MCP transport。
- 面向自动化的稳定 JSON 输出：`--json` 下 stdout 始终是 JSON envelope。
- 成功和失败都可由程序解析：失败时 stdout 仍包含 `error` 对象。
- stderr 只输出诊断、重试、调试等信息，不作为结构化数据来源。
- 支持命令自省：`yunxiao commands --json` 和 `yunxiao <cmd> --help --json`。
- 支持常用只读域：`org`、`codeup`、`flow`、`projex`、`packages`、`testhub`。
- 提供受限 raw request：仅支持 `GET /oapi/...`，用于只读 API 覆盖缺口。
- 内置 timeout、retry、exit code、错误分类和分页契约。

## 安装方法

### 从 Release 下载

项目配置了 GoReleaser，发布 tag 形如 `v*` 时会构建以下平台产物：

- Linux amd64
- Linux arm64
- macOS amd64
- macOS arm64

下载对应平台的压缩包后解压，将 `yunxiao` 放入 `PATH` 中即可：

```bash
tar -xzf yunxiao_<version>_<os>_<arch>.tar.gz
chmod +x yunxiao
mv yunxiao /usr/local/bin/yunxiao
```

然后验证：

```bash
yunxiao commands --json
```

### 从源码安装

需要 Go 1.25.4 或与项目兼容的 Go 版本。

```bash
git clone <repo-url>
cd yunxiao-cli
go build -o yunxiao ./cmd/yunxiao
./yunxiao commands --json
```

如果希望安装到当前 Go 环境的 bin 目录：

```bash
go install ./cmd/yunxiao
```

## 认证与配置

### 获取云效个人访问令牌

使用 CLI 前需要先创建阿里云云效个人访问令牌。可以参考阿里云文档：[获取个人访问令牌](https://help.aliyun.com/zh/yunxiao/developer-reference/obtain-personal-access-token)。创建令牌时，建议授予组织管理、项目协作、代码管理、流水线、制品仓库、应用交付、测试管理下 API 的读写权限，并将令牌到期时间设置为较长期限，避免自动化任务因令牌过期中断。

### 配置访问令牌

人类用户可以直接运行 `yunxiao auth` 进入交互式输入界面。CLI 会在终端中提示输入云效访问令牌，输入内容会明文显示以便核对；请在私密终端中输入，默认验证成功后写入本机配置：

```bash
yunxiao auth
```

如果已有本地 token，需要显式允许覆盖：

```bash
yunxiao auth --force
```

离线或内网环境无法访问验证接口时，可以显式跳过验证；此路径会保存未验证 token，后续业务命令可能仍因 token 无效而失败：

```bash
yunxiao auth --skip-verify --force
```

脚本中非交互写入 token 时，使用 stdin，避免 token 进入 shell history 或进程列表：

```bash
printf '%s\n' "$YUNXIAO_ACCESS_TOKEN" | yunxiao auth login --token-stdin --force --json
```

查看和清除当前认证状态：

```bash
yunxiao auth status --json
yunxiao auth status --verify --json
yunxiao auth logout --json
```

自动化、CI 和 AI Agent 仍推荐直接使用环境变量：

```bash
export YUNXIAO_ACCESS_TOKEN=<your-access-token>
```

`YUNXIAO_ACCESS_TOKEN` 的优先级高于本地配置。`auth logout` 只删除配置文件里的 `access_token`，不会影响当前 shell 中的环境变量。

本地 token 默认保存在 `~/.yunxiao/config.yaml` 的 `access_token` 字段。CLI 写入时会尽力设置目录权限为 `0700`、配置文件权限为 `0600`；不要把该配置文件提交到 git。当前阶段不承诺 Windows 上完整 ACL 修复，如果需要更强隔离，后续可扩展系统 keychain。

常用配置项：

| 参数 / 环境变量 | 说明 | 默认值 |
|---|---|---|
| `--organization-id` / `YUNXIAO_ORGANIZATION_ID` | 云效组织 ID | 无 |
| `--region` / `YUNXIAO_REGION` | 区域选择；当前主要保留为全局参数，实际端点优先由 `YUNXIAO_API_BASE_URL` 决定 | 无 |
| `--timeout` / `YUNXIAO_TIMEOUT` | 请求超时时间，单位秒 | `30` |
| `YUNXIAO_ACCESS_TOKEN` | 云效访问令牌；优先级高于配置文件 | 无 |
| `YUNXIAO_API_BASE_URL` | API 基础地址；中心站默认使用 OpenAPI 域名，Region 站请配置组织专属域名 | `https://openapi-rdc.aliyuncs.com` |
| `YUNXIAO_REGION_DEFAULT_ORG_ID` | 区域版 endpoint 下的组织 ID 环境回退 | 无 |

优先级从高到低：

1. 命令行参数
2. 环境变量（`YUNXIAO_` 前缀）
3. 配置文件 `~/.yunxiao/config.yaml`
4. 内置默认值

区域版 endpoint 下，如果没有显式传入 `--organization-id` 且没有 `YUNXIAO_ORGANIZATION_ID`，CLI 还会尝试使用 `YUNXIAO_REGION_DEFAULT_ORG_ID` 作为组织 ID 回退。

## 基本使用

### 查看当前用户和组织

```bash
yunxiao org current --json
yunxiao org members list --organization-id <org-id> --json
```

### 查看 Codeup 仓库

```bash
yunxiao codeup repos list --organization-id <org-id> --json
yunxiao codeup repo get --organization-id <org-id> --repo-id <repo-id> --json
yunxiao codeup branches list --organization-id <org-id> --repo-id <repo-id> --json
yunxiao codeup commits list --organization-id <org-id> --repo-id <repo-id> --json
yunxiao codeup file get --organization-id <org-id> --repo-id <repo-id> --path <file-path> --ref <ref> --json
yunxiao codeup compare get --organization-id <org-id> --repo-id <repo-id> --from <ref> --to <ref> --json
```

### 查看流水线

```bash
yunxiao flow pipelines list --organization-id <org-id> --json
yunxiao flow pipeline get --organization-id <org-id> --pipeline-id <pipeline-id> --json
yunxiao flow runs list --organization-id <org-id> --pipeline-id <pipeline-id> --json
yunxiao flow run get --organization-id <org-id> --pipeline-id <pipeline-id> --run-id <run-id> --json
```

### 查看 Projex 项目和工作项

```bash
yunxiao projex projects list --organization-id <org-id> --json
yunxiao projex project get --organization-id <org-id> --project-id <project-id> --json
yunxiao projex workitems list --organization-id <org-id> --category <category> --space-id <space-id> --json
yunxiao projex workitem get --organization-id <org-id> --workitem-id <workitem-id> --json
yunxiao projex sprints list --organization-id <org-id> --project-id <project-id> --json
# 可选：通过 --status <status-list> 过滤迭代状态
```

Projex 当前提供项目/空间内列表枚举和已知 ID 详情查询。`workitems list` 仍需要明确的 `--space-id` 和 `--category`，并支持与 MCP server 对齐的 `--status`、`--assigned-to`、`--finish-time-after` 等查询条件；跨项目个人待办、“未完成”业务语义和“本周完成”聚合查询尚未成为公共命令契约。

### 查看制品仓库和制品

```bash
yunxiao packages repos list --organization-id <org-id> --json
yunxiao packages artifacts list --organization-id <org-id> --repo-id <repo-id> --repo-type <repo-type> --json
yunxiao packages artifact get --organization-id <org-id> --repo-id <repo-id> --artifact-id <artifact-id> --repo-type <repo-type> --json
```

### 查看 Testhub 数据

```bash
yunxiao testhub testcases list --organization-id <org-id> --test-repo-id <test-repo-id> --json
yunxiao testhub testcase get --organization-id <org-id> --test-repo-id <test-repo-id> --testcase-id <testcase-id> --json
yunxiao testhub directories list --organization-id <org-id> --test-repo-id <test-repo-id> --json
yunxiao testhub testplans list --organization-id <org-id> --json
```

注意：`testhub testplans list` 当前不支持分页参数，不要传 `--page-size` 或 `--page-token`。

### 只读 raw request

当某个只读 API 暂无封装命令时，可以使用 raw request：

```bash
yunxiao raw request --method GET --path /oapi/v1/custom/resource?foo=1 --json
```

限制：

- 仅支持 `GET`。
- `--path` 必须以 `/oapi/` 开头。
- 不支持绝对 URL。
- 不能用于创建、更新、删除、触发等写操作。

## 命令自省

列出所有命令及参数：

```bash
yunxiao commands --json
```

查看单个命令的结构化帮助：

```bash
yunxiao flow pipeline get --help --json
```

建议 AI Agent 和脚本优先使用自省能力发现命令和参数，不要猜测命令别名或参数名。

## JSON 输出契约

`--json` 输出统一使用 envelope：

```json
{
  "version": "v1",
  "data": null,
  "meta": {},
  "error": null
}
```

成功时：

- `data` 为结果数据。
- `error` 为 `null`。
- exit code 为 `0`。

失败时：

- `data` 为 `null`。
- `error` 包含 `code`、`category`、`retryable`、`message`、`upstream_status`；其中 `upstream_status` 在非 HTTP 错误场景下可能为 `null`。
- stdout 仍然输出完整 JSON envelope。
- stderr 只输出诊断信息。

脚本中应解析 stdout，而不是从 stderr 中提取结构化错误。

示例：

```bash
out=$(yunxiao codeup repos list --organization-id "$ORG_ID" --json 2>diag)
status=$?

if [ "$status" -eq 0 ]; then
  jq '.data' <<<"$out"
else
  jq '.error' <<<"$out" >&2
  exit "$status"
fi
```

## Exit code 与错误分类

| Exit code | 分类 | 含义 |
|---:|---|---|
| 0 | - | 成功 |
| 1 | `general` | 通用失败 |
| 2 | `param` | 参数错误 |
| 3 | `auth` | 认证失败或 token 缺失 |
| 4 | `not_found` | 资源不存在 |
| 5 | `rate_limit` | 触发限流 |
| 6 | `network` | 网络错误、DNS、连接或超时 |
| 7 | `upstream` | 上游服务 5xx |
| 8 | `forbidden` | 权限不足 |

处理建议：

- `param`：修正命令或参数，不要盲目重试。
- `auth`：检查或刷新 `YUNXIAO_ACCESS_TOKEN`。
- `rate_limit`：退避后重试。
- `network`：可对只读命令做有限重试。
- `upstream`：查看 `error.retryable`，注意内置 retry 可能已经耗尽。
- `forbidden`：检查权限，不要盲目重试。

## 分页使用

分页 list 命令支持：

- `--page-size <正整数>`
- `--page-token <token>`

响应中通过 `meta.pagination` 返回分页信息：

```json
{
  "next_token": "token-or-null",
  "page_size": 20,
  "has_more": true
}
```

循环示例：

```bash
token=""
while :; do
  args=(codeup repos list --organization-id "$ORG_ID" --page-size 100 --json)
  [ -n "$token" ] && args+=(--page-token "$token")

  out=$(yunxiao "${args[@]}") || exit $?
  jq -c '.data[]' <<<"$out"

  has_more=$(jq -r '.meta.pagination.has_more // false' <<<"$out")
  [ "$has_more" = true ] || break

  token=$(jq -r '.meta.pagination.next_token // empty' <<<"$out")
  [ -n "$token" ] || break
done
```

注意：`--page-size` 必须是正整数。`0`、负数或非整数会在鉴权前返回 `PARAM_INVALID`。

## 自行开发与编译

### 安装依赖

```bash
go mod download
```

### 运行测试

```bash
go test ./...
```

### 运行格式检查

```bash
test -z "$(gofmt -l cmd internal test)"
```

### 运行 vet

```bash
go vet ./...
```

### 本地构建

```bash
go build -o yunxiao ./cmd/yunxiao
```

### 多平台构建

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/yunxiao-linux-amd64 ./cmd/yunxiao
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o /tmp/yunxiao-linux-arm64 ./cmd/yunxiao
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/yunxiao-darwin-amd64 ./cmd/yunxiao
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o /tmp/yunxiao-darwin-arm64 ./cmd/yunxiao
```

### 推荐本地验证命令

```bash
test -z "$(gofmt -l cmd internal test)" && \
  go vet ./... && \
  go test ./... && \
  go build -o yunxiao ./cmd/yunxiao
```

## CI 与发布

项目包含两个 GitHub Actions workflow：

- `.github/workflows/ci.yml`
  - 在 pull request 和 `main` 分支 push 时运行。
  - 执行 gofmt 检查、`go vet ./...`、`go test ./...` 和 `go build`。

- `.github/workflows/release.yml`
  - 在推送 `v*` tag 时运行。
  - 执行 `go vet ./...`、`go test ./...`，然后使用 GoReleaser 发布。

GoReleaser 配置位于 `.goreleaser.yaml`，当前构建 Linux/macOS 的 amd64/arm64 tar.gz 包，并生成 `checksums.txt`。

## AI Agent 使用建议

本项目提供配套 Skill：

```text
skill/using-yunxiao-cli/SKILL.md
```

建议 Agent 安装或读取该 Skill 后再调用 `yunxiao`。该 Skill 总结了命令发现、stdout JSON envelope 解析、错误分类、分页、raw request 边界等 Agent 常见误用点。

Agent 调用时应遵循：

1. 先用 `yunxiao commands --json` 或 `--help --json` 发现命令。
2. 始终传 `--json`。
3. 始终解析 stdout envelope。
4. 不从 stderr 提取结构化结果。
5. 不猜测未暴露的参数或命令别名。
6. 不使用 `raw request` 做写操作。

## 注意事项

- 当前命令以只读能力为主，写操作尚未作为公共契约提供。
- `raw request` 不是通用 HTTP 客户端，只是只读逃生口。
- `testhub testplans list` 不支持分页参数。
- stdout/stderr 分离是公共契约，不要把 stderr 当成数据源。
- `--quiet` 只影响 stderr 诊断，不改变 stdout JSON envelope。
- token、Authorization header 等敏感信息不应出现在日志和 stderr 中。
- `YUNXIAO_API_BASE_URL` 会优先于 region 相关配置。
- 命令、flag、exit code、JSON 顶层结构属于公共 API，修改前应同步更新测试和契约文档。

## 相关文档

- `docs/adr/001-cli-contract.md`：CLI 公共契约 ADR。
- `docs/phase0-status.md`：Phase 0 状态记录。
- `skill/using-yunxiao-cli/SKILL.md`：AI Agent 使用 Yunxiao CLI 的配套 Skill。
