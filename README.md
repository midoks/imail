<p align="center">
  <img alt="logo" src="https://avatars2.githubusercontent.com/u/4169529?v=3&s=200" height="140" />
  <h3 align="center">imail</h3>
  <p align="center">imail 是一款极易搭建的自助邮件服务。</p>
</p>


---
## 项目愿景

imail项目旨在打造一个以最简便的方式搭建简单、稳定的邮件服务。使用 Go 语言开发使得 imail 能够通过独立的二进制分发，并且支持 Go 语言支持的 所有平台，包括 Linux、macOS、Windows 以及 ARM 平台。

- 支持多域名管理。
- 邮件草稿功能支持。
- 邮件搜索功能支持。
- Rspamd垃圾邮件过滤支持。
- Hook脚本支持。

[![Go](https://github.com/midoks/imail/actions/workflows/go.yml/badge.svg)](https://github.com/midoks/imail/actions/workflows/go.yml)
[![CodeQL](https://github.com/midoks/imail/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/midoks/imail/actions/workflows/codeql-analysis.yml)
[![Codecov](https://codecov.io/gh/midoks/imail/branch/master/graph/badge.svg?token=MJ2HL6HFLR)](https://codecov.io/gh/midoks/imail)

## 版本截图

[![main](/screenshot/main.png)](/screenshot/main.png)


## 版本详情

- 0.0.17

```
* 增加修改管理员修改密码功能.
* 优化日志显示.
* initd 改为systemd.
* 修复初始化无法登录的现象.
* ssl功能优化.
```

## Wiki

- https://github.com/midoks/imail/wiki

## 贡献者

[![](https://contrib.rocks/image?repo=midoks/imail)](https://github.com/midoks/imail/graphs/contributors)


## 授权许可

本项目采用 MIT 开源授权许可证，完整的授权说明已放置在 [LICENSE](https://github.com/midoks/imail/blob/main/LICENSE) 文件中。

