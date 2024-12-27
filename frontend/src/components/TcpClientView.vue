<template>
  <div  :style="{ width: windowWidth + 'px', height: windowHeight + 'px' }" class="tcp-client-view">
    <div class="header">
      <div class="connection-info">
        <span class="id">{{ id }}</span>
        <span class="address">{{ address }}</span>
        <span class="status-badge" :class="connectionStatusClass">
          {{ connectionStatusText }}
        </span>
      </div>
      <div class="actions">
        <button class="btn btn-edit" @click="handleEdit(clientData.id)">编辑</button>
        <button 
          class="btn" 
          :class="isConnected ? 'btn-danger' : 'btn-primary'"
          @click="handleConnection"
          :disabled="connecting"
        >
          {{ connectionButtonText }}
        </button>
        <button class="btn" @click="clearMessages(clientData.id)">清空消息</button>
      </div>
    </div>

    <div class="message-area">
      <div v-if="messages.length === 0" class="no-data">
        <p>没有数据</p>
      </div>
      <div v-else class="messages-container">
        <div v-for="msg in messages" 
             :key="msg.id" 
             class="message-item"
             :class="{ 
               'message-outgoing': msg.direction === 'outgoing',
               'message-incoming': msg.direction === 'incoming'
             }">
          <div class="message-content">
            <div class="message-text">{{ msg.content }}</div>
            <div class="message-time">{{ formatTime(msg.timestamp) }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="input-panel">
      <div class="input-controls">
        <span>输入方式</span>
        <select class="select-control" v-model="inputMethod">
          <option value="text">文本</option>
          <option value="hex">16进制</option>
        </select>

        <span>显示方式</span>
        <select class="select-control" v-model="displayMethod">
          <option value="text">文本</option>
          <option value="hex">16进制</option>
        </select>

        <span>编码</span>
        <select class="select-control" v-model="encoding">
          <option value="utf8">UTF-8</option>
        </select>
      </div>
      <div class="message-input">
        <input v-model="inputContent" type="text" placeholder="输入要发送的文本，Command+回车(⌘ + ↩)换行" />
        <button class="btn btn-primary" :class="sendScheduledClass" @click="sendMessage">{{ sendScheduledText }}</button>
      </div>
    </div>

    <AddTcpClientModal 
      :visible="isEditModalVisible" 
      :onClose="closeEditModal"
      :editMode="true"
      :initialData="editData"
      @refresh-list="handleRefresh"
    />
  </div>
</template>

<script>
import AddTcpClientModal from './AddTcpClientModal.vue'
import { GetTCPClientData, ConnectTCPClient, DisconnectTCPClient, GetTCPClientStatus, SendMessage, SendScheduledMessage, StopScheduledMessage } from "../../wailsjs/go/control/FuncTcpClient"
import { DeleteMessage, GetAllMessages } from "../../wailsjs/go/control/Message"
import { ShowWarningDialog } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
// 导入按钮样式
export default {
  name: 'TcpClientView',
  components: {
    AddTcpClientModal
  },
  props: {
    id: {
      type: [String, Number],
      required: true
    },
    address: {
      type: String,
      required: true
    },
    clientData: {
      type: Object,
      required: true
    }
  },
  // 数据
  data() {
    return {
      hasMessages: false,
      isEditModalVisible: false,
      editData: null,
      messages: [],
      connecting: false,
      sendScheduled: false,
      isConnected: false,
      windowWidth: window.innerWidth - 250 - 250, // 当前窗口宽度减去列表和侧边栏宽度
      windowHeight: window.innerHeight, // 当前窗口高度
      inputContent: '', // 新增输入内容的绑定
      inputMethod: 'text', // 输入方式
      displayMethod: 'text', // 显示方式
      encoding: 'utf8', // 编码
    }
  },
  // 计算属性
  computed: {
    connectionStatusClass() {
      if (this.connecting) return 'status-connecting'
      return this.isConnected ? 'status-connected' : 'status-disconnected'
    },
    connectionStatusText() {
      if (this.connecting) return '连接中...'
      return this.isConnected ? '已连接' : '未连接'
    },
    connectionButtonText() {
      if (this.connecting) return '断开连接中...'
      return this.isConnected ? '断开' : '连接'
    },
    
    connectionButtonClass() {
        return this.clientData.status === 'online' ? 'btn-primary' : 'btn-danger'
    },
    sendScheduledClass() {
      return this.sendScheduled ? 'btn-primary' : 'btn-danger'
    },
    sendScheduledText() {
      return this.sendScheduled ? '停止' : '发送'
    } 
  },
  // 生命周期
  mounted() {
    this.loadMessages()
    this.checkConnectionStatus()
    // 监听窗口大小变化
    window.addEventListener('resize', this.handleResize);
    // 监听事件
    EventsOn('client_event', (event) => {
      // 检查事件是否与当前客户端相关
      if (event.server_id == this.clientData.id) {
        if (event.type == 'data_sent') {
          console.log(event.message)
          this.messages.push(event.message)
          this.$nextTick(() => {
            this.scrollToBottom()
          })
        } else if (event.type == 'data_received') {
          console.log(event.message)

          this.messages.push(event.message)
          this.$nextTick(() => {
            this.scrollToBottom()
          })
        }
      }
    })
  },

  // 方法
  methods: {
     handleResize() {
      this.windowWidth = window.innerWidth - 250 - 250; // 更新窗口宽度，减去列表和侧边栏宽度
      this.windowHeight = window.innerHeight; // 更新窗口高度
    },
    // 检查连接状态
    async checkConnectionStatus() {
      try {
        const response = await GetTCPClientStatus(this.clientData.id)
        if (response && response.data) {
          this.isConnected = response.data.status === 'online'
        } 
      } catch (error) {
        window.runtime.LogError('获取连接状态失败: ' + error)
      }
    },
    // 连接或断开
    async handleConnection() {
      try {
        if (this.connecting) return
        this.connecting = true
        if (this.isConnected) {
          const response = await DisconnectTCPClient(this.clientData.id)
          if (response && response.success) {
            this.isConnected = false
            this.sendScheduled = false
            await StopScheduledMessage(this.clientData.id)
            window.runtime.LogInfo('已断开连接')
          } else {
            window.runtime.LogError('断开连接失败')
          }
        } else {
          this.connecting = true
          const response = await ConnectTCPClient(this.clientData.id) 
          if (response && response.success) {
            this.isConnected = true
            window.runtime.LogInfo('连接成功')
          } else {
            window.runtime.LogError('连接失败')
          }
        }
      } catch (error) {
        window.runtime.LogError('连接操作失败: ' + error)
      } finally {
        this.connecting = false
      }
    },
    // 加载消息
    async loadMessages() {
      try {
        const response = await GetAllMessages(this.clientData.id)
        if (response && response.data) {
          this.messages = response.data
          this.hasMessages = this.messages.length > 0
          // 滚动到最新消息
          this.$nextTick(() => {
            this.scrollToBottom()
          })
        }
      } catch (error) {
        window.runtime.LogError('获取消息记录失败: ' + error)
      }
    },
    // 格式化时间
    formatTime(timestamp) {
      const date = new Date(timestamp)
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit', 
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      })
    },
    // 滚动到最新消息
    scrollToBottom() {
      const container = document.querySelector('.messages-container')
      if (container) {
        container.scrollTop = container.scrollHeight
      }
    },
    // 编辑
    async handleEdit(id) {
      try {
        const response = await GetTCPClientData(id)
        if (response && response.data) {
          this.editData = response.data
          this.isEditModalVisible = true
        } else {
          window.runtime.LogError('获取数据失败')
        }
      } catch (error) {
        window.runtime.LogError('获取数据出错: ' + error)
      }
    },
    // 关闭编辑模态框
    closeEditModal() {
      this.isEditModalVisible = false
      this.editData = null
      this.$emit('refresh-list')
    },
    // 刷新列表
    handleRefresh() {
      this.$emit('refresh-list')
    },
    // 停止定时发送
    async stopScheduledMessage() {
      await StopScheduledMessage(this.clientData.id)
      window.runtime.LogInfo('定时发送任务已停止')
    },
    // 发送消息
    async sendMessage() {
      if (!this.inputContent && !this.sendScheduled) {
        window.runtime.LogError('输入内容不能为空')
        await ShowWarningDialog('警告', "输入内容不能为空")
        return
      }

      if (!this.isConnected) {
        window.runtime.LogError('未连接')
        await ShowWarningDialog('警告', "未连接")
        return
      }

      try {
        if (this.sendScheduled) {
          this.sendScheduled = false
          await StopScheduledMessage(this.clientData.id)
        } else {
            if (this.clientData.repeatSend) {
              await SendScheduledMessage(this.clientData.id, this.inputContent, this.inputMethod, this.clientData.repeatInterval)
              this.sendScheduled = true
            } else {
              await SendMessage(this.clientData.id, this.inputContent, this.inputMethod)
            }
            this.inputContent = '' // 清空输入框
          this.loadMessages() // 重新加载消息
        }
      } catch (error) {
        window.runtime.LogError('发送消息出错: ' + error)
      }
    },
  // 清空消息
  async clearMessages(id) {
    this.messages = []
    await DeleteMessage(id)
    this.loadMessages()
  },
  },

  // 销毁
  beforeDestroy() {
    EventsOff('client_event')
  }
}
</script>

<style scoped>
.tcp-client-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: #ffffff;
  color: #333333;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e8e8e8;
}

.connection-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.actions {
  display: flex;
  gap: 8px;
}

/* 统一按钮样式 */
.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  transition: all 0.3s;
  background-color: #f0f0f0;
  color: #333333;
}

.btn:hover {
  background-color: #e0e0e0;
}

.btn-primary {
  background-color: #1890ff;
  color: white;
}

.btn-primary:hover {
  background-color: #40a9ff;
}

.message-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: #ffffff;
  overflow: hidden;
  position: relative;
}

.no-data {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  color: #999;
}

.no-data img {
  width: 64px;
  height: 64px;
  margin-bottom: 10px;
}

.input-panel {
  padding: 16px 20px;
  border-top: 1px solid #e8e8e8;
}

.input-controls {
  display: flex;
  gap: 16px;
  align-items: center;
  margin-bottom: 16px;
}

/* 统一下拉框样式 */
.select-control {
  padding: 6px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  background-color: #ffffff;
  color: #333333;
  font-size: 14px;
}

.select-control:hover {
  border-color: #40a9ff;
}

.select-control:focus {
  border-color: #1890ff;
  outline: none;
  box-shadow: 0 0 0 2px rgba(24, 144, 255, 0.2);
}

.message-input {
  display: flex;
  gap: 8px;
}

.message-input input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  color: #333333;
  font-size: 14px;
}

.message-input input:focus {
  border-color: #1890ff;
  outline: none;
  box-shadow: 0 0 0 2px rgba(24, 144, 255, 0.2);
}

.message-input input::placeholder {
  color: #999;
}

.messages-container {
  padding: 20px;
  overflow-y: auto;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.message-item {
  display: flex;
  margin-bottom: 10px;
  max-width: 70%;
}

.message-incoming {
  align-self: flex-start;
}

.message-outgoing {
  align-self: flex-end;
}

.message-content {
  padding: 10px 15px;
  border-radius: 8px;
  position: relative;
}

.message-incoming .message-content {
  background-color: #f0f0f0;
  margin-right: auto;
}

.message-outgoing .message-content {
  background-color: #1890ff;
  color: white;
  margin-left: auto;
}

.message-text {
  word-break: break-word;
  font-size: 14px;
  line-height: 1.4;
}

.message-time {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
  text-align: right;
}

.message-outgoing .message-time {
  color: rgba(255, 255, 255, 0.8);
}

.messages-container::-webkit-scrollbar {
  width: 6px;
}

.messages-container::-webkit-scrollbar-track {
  background: transparent;
}

.messages-container::-webkit-scrollbar-thumb {
  background-color: rgba(0, 0, 0, 0.2);
  border-radius: 3px;
}

.messages-container::-webkit-scrollbar-thumb:hover {
  background-color: rgba(0, 0, 0, 0.3);
}
</style> 