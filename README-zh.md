torsniff - 一个 BitTorrent 网络种子嗅探器
======================================

**简体中文** | [English](./README.md)

## 介绍

torsniff 是一个种子嗅探器，它从 BitTorrent 网络获取人们下载音乐、电影、游戏、文档等等时所用的种子。

种子含有可用的信息，所以你可用 torsniff 打造属于自己的私人种子库（比如：海盗湾），或者做数据挖掘与分析。

## 安装

直接从 [releases](https://github.com/fanpei91/torsniff/releases) 下载最新版即可. 

## 用法

```
$ ./torsniff -h

Usage:
  torsniff [flags]

Flags:
  -a, --addr string        listen on given address (default all, ipv4 and ipv6)
  -h, --help               help for torsniff
  -f, --friends int        max fiends to make with per second (default 500)
  -e, --peers int          max peers to connect to download torrents (default 400)
  -p, --port uint16        listen on given port (default 6881)
  -t, --timeout duration   max time allowed for downloading torrents (default 10s)
  -v, --verbose            run in verbose mode (default true)
  -H, --http-port int      HTTP server port (default 8090)
  -r, --max-retries int    maximum number of retries to fetch metadata (default 3)
```

## 快速开始
使用默认参数:

`./torsniff`

种子默认保存在 `$HOME/torrents` 目录里。

## 环境要求
* 需要一个有公网 IP 的主机（推荐，最好是国外），如果想在私有内网、NAT 内的主机上运行，需要配置端口转发、映射。
* 允许 UDP 流量通过防火墙
* 你的 ISP/主机商允许 BitTorrent 流量（torsniff 在 [vultr.com](https://www.vultr.com/?ref=7172229) 能良好运行）

### UPnP 支持

该应用程序现在支持 UPnP（通用即插即用），可以在兼容的路由器上自动配置端口转发。这一功能有助于在 NAT 或防火墙后运行应用程序时更轻松地建立连接。

## 前端设置

要设置前端，请导航到 `frontend` 目录并运行：

```bash
npm install
npm run build
```

这将构建前端资产并将其放置在 `../static` 目录中。

## 协议
- [DHT Protocol](http://www.bittorrent.org/beps/bep_0005.html)
- [The BitTorrent Protocol Specification](http://www.bittorrent.org/beps/bep_0003.html)
- [BitTorrent  Extension Protocol](http://www.bittorrent.org/beps/bep_0010.html)
- [Extension for Peers to Send Metadata Files](http://www.bittorrent.org/beps/bep_0009.html)

## 许可证
MIT
