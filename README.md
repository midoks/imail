<p align="center">
  <img alt="imail logo" src="https://avatars2.githubusercontent.com/u/4169529?v=3&s=200" height="140" />
  <h3 align="center">imail</h3>
  <p align="center">imail 是一款极易搭建的自助邮件服务。</p>
</p>

---
## 项目愿景

imail项目旨在打造一个以最简便的方式搭建简单、稳定的邮件服务。使用 Go 语言开发使得 imail 能够通过独立的二进制分发，并且支持 Go 语言支持的 所有平台，包括 Linux、macOS、Windows 以及 ARM 平台。


## 快速入口
- [文档主页](https://github.com/midoks/imail/wiki)
- [API](https://github.com/midoks/imail/wiki/API%E6%96%87%E6%A1%A3)

### 版本详情

0.0.4

- 添加支持sqlite - [xmapst](https://github.com/xmapst)
- 添加API白名单 - [xmapst](https://github.com/xmapst)
- 优化redis连接 - [xmapst](https://github.com/xmapst)
- 解决邮件移动问题[exp:收件箱->垃圾箱->收件箱]

### 待解决问题
- [ ] 前端需求
- [ ] [系统打包](https://github.com/midoks/imail/issues/5)
- [ ] test不完整
- [ ] 日志配置功能
- [ ] ...


## 计划功能

- [x] 邮件接收功能[POP3,IMAP,SMTP]
- [x] 邮件投递功能[SMTP]
- [x] dkim && check
- [x] rspamd
- [x] hook脚本支持
- [ ] API
- [ ] 性能优化

## 快速安装

```
curl -fsSL  https://raw.githubusercontent.com/midoks/imail/master/scripts/install.sh | sh
```

## 快速开发
```
curl -fsSL  https://raw.githubusercontent.com/midoks/imail/master/scripts/install_dev.sh | sh
```

## 前端项目

计划开发[imail前端](https://github.com/midoks/imail-web),由于精力有限。建立的这个项目，不知道有机会写不～。

## 授权许可

本项目采用 MIT 开源授权许可证，完整的授权说明已放置在 [LICENSE](https://github.com/midoks/imail/blob/main/LICENSE) 文件中。

