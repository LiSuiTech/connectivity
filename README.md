# 网络调试工具

## 项目简介

网络调试工具是一个基于 Go 和 Web 技术构建的桌面应用程序。它提供了一个集成的环境，用于调试和测试网络连接，包括 TCP 和 UDP 客户端和服务器功能。

## 功能特性

- **TCP/UDP 客户端和服务器**：支持创建和管理多个 TCP 和 UDP 连接。
- **消息发送与接收**：支持文本和十六进制格式的消息发送与接收。
- **连接状态监控**：实时监控连接状态，并提供连接管理功能。
- **跨平台支持**：支持 Windows、macOS 和 Linux。

## 安装

请确保您的系统上已安装 Go 和 Node.js。

1. 克隆项目到本地：
   ```bash
   git clone https://github.com/YouEvanLi/connectivity.git
   cd <repository-directory>
   ```

2. 安装前端依赖：
   ```bash
   npm install
   ```

3. 构建前端：
   ```bash
   npm run build
   ```

4. 运行应用程序：
   ```bash
   wails dev
   ```

## 使用说明

- 启动应用后，您可以在侧边栏选择不同的功能模块。
- 在 TCP/UDP 客户端或服务器模块中，您可以创建新的连接、发送和接收消息。
- 使用顶部的菜单栏可以访问更多设置和选项。

## 开发

### 开发环境

- **推荐 IDE**：Visual Studio Code
- **插件**：Volar（用于 Vue 3 支持）

### 运行开发模式

在项目目录下运行以下命令以启动开发模式：

```bash
wails dev
```

这将启动一个 Vite 开发服务器，提供快速的前端热重载。

### 构建生产版本

使用以下命令构建可分发的生产模式包：

```bash
wails build
```

## 贡献

欢迎贡献代码！请提交 Pull Request 或报告问题。

## 许可证

该项目基于 MIT 许可证进行分发。详情请参阅 LICENSE 文件。
