[English](PULL_REQUEST_TEMPLATE.md) | [简体中文](PULL_REQUEST_TEMPLATE.zh-CN.md) | [Website / 官网](https://dotapk.lol)

# Pull request

## 问题与修改后行为

说明具体触发条件与对用户/开发者的变化。

## 验证

列实际命令和证明范围，区分源码/模块测试与真实浏览器/网络验收。

- [ ] 中英指南同时更新，首部语言/官网和相对链接正确。
- [ ] `python3 scripts/check-docs.py` 与 `git diff --check` 通过。
- [ ] 适用焦点检查串行完成，未验证限制说明。
- [ ] 稳定ID、22/88发布与暂停状态保持，或有独立接受范围说明变化。
- [ ] 无凭据/私有路径/逐玩家或逐局报告/可追踪QA数据/未授权媒体。
- [ ] MIT和第三方边界保留，无隐式部署/DB/CI权限变化。

见[贡献](../CONTRIBUTING.zh-CN.md)。标准LICENSE保持英文，中文指南只解释范围不替代法律文本。
