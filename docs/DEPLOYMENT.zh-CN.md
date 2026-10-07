[English](DEPLOYMENT.md) | [简体中文](DEPLOYMENT.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# 部署与运维边界

项目复用现有 AgentSquared MySQL8.4/Nginx，使用独立 `dota_duel` schema。Go 监听127.0.0.1:18082，Nginx 提供 `https://api.dotapk.lol`，Cloudflare 提供 `https://dotapk.lol`。源码维护不新增数据库/服务器、自动平衡同步或公开统计接口。

保留操作者输入：`deploy/api.dotapk.lol.conf`、`deploy/dota-duel.service`、`deploy/public.env`、`deploy/provision-database.py`、`deploy/check-runtime-access.py` 和 [production-v13 profile](../deploy/production-v13/README.zh-CN.md)。public.env 是非秘密运行配置，DSN 放操作者控制的独立私有路径。provision/check 脚本绑定本应用隔离策略，不是通用安装命令，贡献者不要执行或伪造替换凭据；已有公开域名和 loopback 路由是有功能用途的配置。

独立安装需审阅自己的 OS/服务用户、二进制/secret/证书路径、schema权限和精确 origin；文档示例用 `<HOST>`、`<LOCAL_BINARY_PATH>`、`<DSN_FILE>`，不用真实私有主机/IP/用户路径。不要把占位符盲目替入运行生产配置。compose 仅隔离本地/staging，不替换已有服务。

## profile、迁移与发布顺序

普通 source embed 是 legacy20，生产 profile 保留 legacy20、旧46与当前22精确 build，前端只启用22/88。管理角色应用已审核增量迁移/roster SQL，构建匹配 embed/overlay，核对 health/registry/CORS，再发布匹配前端。运行只需 DML，SQL INSERT和精确前置 compare-and-swap append 不能当无条件可重复 provision。见[开发](DEVELOPMENT.zh-CN.md)和 [API](API.zh-CN.md)。

`python3 scripts/build-production-v13.py` 用精确 Go1.26.1 和已有依赖 cache 复现固定 Linux 产物，不访问数据库/服务；检查冻结源字节/profile SHA及二进制 SHA。它证明该历史产物，不证明当前在线或浏览器验收。[profile README](../deploy/production-v13/README.zh-CN.md)列出不可变输入，不削弱检查或重新固定值来隐藏不一致。

回滚要考虑兼容记录/房间：停止/排空 roster-aware 客户端会话，保留可读取已存在扩展记录的版本，保留增量表/view/原报告，不改写旧结果或复用 build ID 表示不同规则。操作者产物/备份放跟踪源码之外；普通清理/许可/文档提交无需部署。

当前源码已移除详细历史部署回执、生产主机/IP笔记与逐局 QA，它们仍在 Git 历史；这是普通文件清理，不是 secret撤销或历史抹除。轻量检查未发现真实凭据，后续发现应私下只报类型/路径，不打印值。
