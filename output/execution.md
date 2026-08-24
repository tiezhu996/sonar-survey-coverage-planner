# gb-532 验证执行记录

- 项目：侧扫声呐测线覆盖缺口规划（`sonar-survey-coverage-planner`）
- 验证日期：2026-08-22（Asia/Shanghai）
- 实现提交：`095aacccda1e4f2501449374502c620f33f9daae`
- Git 身份：`blueship581 <brysj.hhrhl.g@gmail.com>`

## 提示词与实现检查

- 按 `gb-532.md` 完成测区、测线规划、声呐航迹和覆盖缺口四实体的 model、DTO、repository、service、handler 分层。
- 覆盖 JWT、RBAC、request ID、统一错误、状态机、幂等、乐观锁、PostGIS 初始化和不可变审计投影。
- 覆盖计算使用米制投影、结构化 GeoJSON、固定网格近似、输入哈希和人工补测决策；页面明确不连接船舶、声呐或自动驾驶设备。
- 官方规模脚本结果：`2988` 行功能 Go 代码、`42` 个功能 Go 文件、2000 档配额 `10` 条。

## 构建与测试

以下命令均以退出码 0 完成：

```text
go build ./backend/...
go vet ./backend/...
go test ./backend/...
go test -race ./backend/...
npm --prefix frontend test                    # 2 files / 4 tests passed
npm --prefix frontend run typecheck
npm --prefix frontend run build               # Vite 8.2.2, 942 modules
npm --prefix frontend audit --registry=https://registry.npmjs.org
docker compose config --quiet
python3 /Users/gaobo/.codex/skills/go-annotation-pipeline/scripts/runtime_smoke.py .
python3 /Users/gaobo/.codex/skills/go-annotation-pipeline/scripts/project_scale.py .
```

- npm audit：`0 vulnerabilities`。
- runtime smoke：SQLite 服务在 `20532` 实际启动，`/healthz` 返回 HTTP 200。

## Compose 与 PostgreSQL API

- 最终镜像启动后，frontend、backend、PostgreSQL 三服务全部为 `healthy`。
- 宿主端口：frontend `18532`、backend `19532`、PostgreSQL `57532`；未暴露其他服务端口。
- PostgreSQL 场景执行 `38` 次真实 HTTP 请求、`67` 条断言，全部通过。
- 覆盖登录成功和失败、401、403、422、409、四实体列表与详情、测线生成/锁定/复制、航迹 checksum 幂等、`imported -> quality_checked -> processing -> processed` 状态机、覆盖计算幂等、reviewer 两步复核和四实体审计。
- 最终 `/api/healthz` 返回 HTTP 200；容器日志未出现 panic、fatal 或 HTTP 5xx。日志中的 401/403/409/422 为上述预期失败路径。

## 内置 Browser 验证

全程使用 Codex 内置 Browser，未使用 Chrome、Computer Use 或外部 Playwright。

- `/areas`：以 admin 创建 `UI-AREA-532`，EPSG:32650、20 m 分辨率、180 m 默认扫幅，列表和本地几何画布同步更新。
- `/plans`：为该测区生成 `UI-AREA-532 coverage plan`，确认归属后锁定为 V2，画布显示边界和平行测线。
- `/runs`：导入 `UI-RUN-532`，核对 2820 m 轨迹长度和 0.63 m/s 平均速度，并依次完成质量检查、处理启动和已处理状态。
- `/coverage`：冻结 `UI-AREA-532 + UI-RUN-532` 计算快照 #2，得到覆盖率 57.8%、缺口率 42.2%、严重度 critical 和补测线；reviewer 记录独立意见后推进 `detected -> reviewed -> accepted`。
- `/audit`：共 22 条事件，四实体计数为 `2 / 6 / 8 / 6`；最新接受事件 request ID 为 `3e9fdfe1a7e7857d8658343a6d15c486`，前后快照准确显示 `reviewed -> accepted`、版本 `2 -> 3` 和冻结输入哈希。
- 新标签页重新登录并依次访问五个主页面后，`tab.dev.logs()` 为空。
- 在 `390x844` 下验证 `/coverage` 和 `/audit`，两页均为 `clientWidth=scrollWidth=bodyScrollWidth=390`，无页面级横向溢出，核心内容和交互可用。

截图：

```text
output/areas-desktop.jpg
output/plans-desktop.jpg
output/runs-desktop.jpg
output/coverage-desktop.jpg
output/coverage-reviewed-desktop.jpg
output/audit-desktop.jpg
output/audit-detail-desktop.jpg
output/coverage-mobile.jpg
output/audit-mobile.jpg
```

## 停服与清理

执行：

```text
docker compose down -v --remove-orphans
```

- `docker compose ps --all` 为空。
- 项目专属容器、网络和 PostgreSQL 数据卷均已删除。
- `18532`、`19532`、`57532`、`20532` 均无监听进程。
- 未执行全局 prune，也未停止或删除其他项目资源。
