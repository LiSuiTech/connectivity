<template>
  <div class="list-container">
    <!-- 左侧列表部分 -->
    <div class="list-section">
      <div class="list-header">
        <h3>列表</h3>
        <div class="button-container">
          <div class="button-group">
            <button class="btn btn-add" @click="showAddModal">
              <i class="fas fa-plus"></i> 新增
            </button>
            <button class="btn btn-delete" @click="handleDelete" :disabled="!selectedItem">
              <i class="fas fa-trash"></i> 删除
            </button>
          </div>
        </div>
      </div>
      
      <div class="list-content">
        <div 
          v-for="item in items" 
          :key="item.id"
          class="list-item"
          :class="{ 'selected': selectedItem?.id === item.id }"
          @click="selectItem(item)"
        >
          <span style="color: #000000">{{ item.name }}</span>
        </div>
      </div>
    </div>

    <!-- 修改右侧内容部分 -->
    <WelcomePage v-if="!selectedItem" />
    <TcpClientView
      v-else-if="pageType === 'tcp-client'"
      :id="selectedItem.rawData.id"
      :address="`${selectedItem.rawData.host}:${selectedItem.rawData.port}`"
      :clientData="selectedItem.rawData"
      @refresh-list="loadData(pageType)"
    />
    <TcpServerView
      v-else-if="pageType === 'tcp-server'"
      :id="selectedItem.rawData.id"
      :address="`${selectedItem.rawData.host}:${selectedItem.rawData.port}`"
      :serverData="selectedItem.rawData"
      @refresh-list="loadData(pageType)"
    />
    <UdpClientView
      v-else-if="pageType === 'udp-client'"
      :id="selectedItem.rawData.id"
      :address="`${selectedItem.rawData.host}:${selectedItem.rawData.port}`"
      :clientData="selectedItem.rawData"
      @refresh-list="loadData(pageType)"
    />
    <UdpServerView
      v-else-if="pageType === 'udp-server'"
      :id="selectedItem.rawData.id"
      :address="`${selectedItem.rawData.host}:${selectedItem.rawData.port}`"
      :serverData="selectedItem.rawData"
      @refresh-list="loadData(pageType)"
    />
    <div v-else class="content-section">
      <div class="content-detail">
        选中项: {{ selectedItem.name }}
      </div>
    </div>
    <AddTcpClientModal :visible="isModalVisibleTcpClient" :onClose="closeAddModal" @refresh-list="loadData(pageType)" />
    <AddTcpServerModal :visible="isModalVisibleTcpServer" :onClose="closeAddModal" @refresh-list="loadData(pageType)" />
    <AddUdpClientModal :visible="isModalVisibleUdpClient" :onClose="closeAddModal" @refresh-list="loadData(pageType)" />
    <AddUdpServerModal :visible="isModalVisibleUdpServer" :onClose="closeAddModal" @refresh-list="loadData(pageType)" />
  </div>
</template>

<script>
import { GetAllTCPClients,DeleteTCPClient, GetTCPClientData } from '../../wailsjs/go/control/FuncTcpClient'
import { GetAllTCPServers,DeleteTCPServer,GetTCPServerData } from '../../wailsjs/go/control/FuncTcpServer'
import { GetAllUdpClients,DeleteUdpClient,GetUdpClientData } from '../../wailsjs/go/control/FuncUdpClient'
import { GetAllUdpServers,DeleteUdpServer,GetUdpServerData } from '../../wailsjs/go/control/FuncUdpServer'
import WelcomePage from './WelcomePage.vue'
import AddTcpClientModal from './AddTcpClientModal.vue'
import AddTcpServerModal from './AddTcpServerModal.vue'
import TcpClientView from './TcpClientView.vue'
import TcpServerView from './TcpServerView.vue'
import UdpClientView from './UdpClientView.vue'
import UdpServerView from './UdpServerView.vue'
import AddUdpClientModal from './AddUdpClientModal.vue'
import AddUdpServerModal from './AddUdpServerModal.vue'
import { ShowWarningDialog } from '../../wailsjs/go/main/App'
import { DeleteMessage } from '../../wailsjs/go/control/Message'

export default {
  name: 'ListView',
  components: { 
    WelcomePage,
    TcpClientView,
    TcpServerView,
    UdpClientView,
    UdpServerView,
    AddTcpClientModal,
    AddTcpServerModal,
    AddUdpClientModal,
    AddUdpServerModal,
  },
  props: {
    // 接收当前页面类型
    pageType: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      items: [],
      selectedItem: null,
      windowWidth: window.innerWidth, // 当前窗口宽度
      windowHeight: window.innerHeight, // 当前窗口高度
      isModalVisibleTcpClient: false,
      isModalVisibleTcpServer: false,
      isModalVisibleUdpClient: false,
      isModalVisibleUdpServer: false,
    }
  },
  watch: {
    // 监听页面类型变化，重新加载数据
    pageType: {
      immediate: true,
      handler(newType) {
        console.log(newType)
        this.loadData(newType);
      }
    }
  },
  mounted() {
    // 监听窗口大小变化
    window.addEventListener('resize', this.handleResize);
  },
  beforeDestroy() {
    // 移除事件监听器
    window.removeEventListener('resize', this.handleResize);
  },
  methods: {
    handleResize() {
      this.windowWidth = window.innerWidth;
      this.windowHeight = window.innerHeight;
    },
    async loadData(type) {
      try {
        let data = [];
        switch (type) {
          case 'tcp-client':
            data = await GetAllTCPClients();
            break;
          case 'tcp-server':
            data = await GetAllTCPServers();
            break;
          case 'udp-client':
            data = await GetAllUdpClients();
            break;
          case 'udp-server':
            data = await GetAllUdpServers();
            console.log(data)
            break;
          default:
            console.warn('Unknown page type:', type);
            return;
        }
        this.items = this.formatData(data);
        // 对列表进行倒序排序
        this.items.reverse();
      } catch (error) {
        console.error('Error loading data:', error);
        // 可以添加错误提示
      }
    },

    // 格式化数据的辅助方法
    formatData(data) {
        if (data.length === 0) {
            return []
        }
      
        if (data.data == null) {
            return []
        }

        return data.data.map((item, index) => ({
            id: item.id || index,
            name: this.getDisplayName(item),
            rawData: item // 保存完整的原始数据
        }));
    },

    // 根据不同类型的数据生成显示名称
    getDisplayName(item) {
      switch (this.pageType) {
        case 'tcp-client':
          return ` ${item.host}:${item.port}`;
        case 'tcp-server':
          return ` ${item.host}:${item.port}`;
        case 'udp-client':
          return ` ${item.host}:${item.port}`;
        case 'udp-server':
          return ` ${item.host}:${item.port}`;
        case 'ping':
          return item.target;
        case 'port-scanner':
          return `${item.target} (${item.ports})`;
        default:
          return item.name || '未命名';
      }
    },
    async handleDelete() {
      if (!this.selectedItem) return;
      
      try {
        let response = null
        // 根据不同类型调用不同的删除接口
        switch (this.pageType) {
          case 'tcp-client':
            response = await DeleteTCPClient(this.selectedItem.rawData.id)
            if (response.success) {
              // 从列表中移除
              await DeleteMessage(this.selectedItem.rawData.id)
              this.items = this.items.filter(item => item.id !== this.selectedItem.id)
              this.selectedItem = null
              window.runtime.LogInfo('删除成功')
            } else {
              await ShowWarningDialog('警告', response.message || '删除失败')
            }
            break;
          case 'tcp-server':
            response = await DeleteTCPServer(this.selectedItem.rawData.id)
            if (response.success) {
              // 从列表中移除
              await DeleteMessage(this.selectedItem.rawData.id)
              this.items = this.items.filter(item => item.id !== this.selectedItem.id)
              this.selectedItem = null
              window.runtime.LogInfo('删除成功')
            } else {
              await ShowWarningDialog('警告', response.message || '删除失败')
            }
            break;
          case 'udp-client':
            response = await DeleteUdpClient(this.selectedItem.rawData.id)
            if (response.success) {
              // 从列表中移除
              await DeleteMessage(this.selectedItem.rawData.id)
              this.items = this.items.filter(item => item.id !== this.selectedItem.id)
              this.selectedItem = null
              window.runtime.LogInfo('删除成功')
            } else {
              await ShowWarningDialog('警告', response.message || '删除失败')
            }
            break;
          case 'udp-server':
            response = await DeleteUdpServer(this.selectedItem.rawData.id)
            if (response.success) {
              // 从列表中移除
              await DeleteMessage(this.selectedItem.rawData.id)
              this.items = this.items.filter(item => item.id !== this.selectedItem.id)
              this.selectedItem = null
              window.runtime.LogInfo('删除成功')
            } else {
              await ShowWarningDialog('警告', response.message || '删除失败')
            } 
            break;
          // ... 其他类型的删除处理
        }
      } catch (error) {
        window.runtime.LogError('删除失败: ' + error)
        await ShowWarningDialog('错误', '删除失败: ' + error)
      }
    },

    async selectItem(item) {
      try {
        console.log(this.pageType)
        let response = null 
        // 根据不同类型获取详细数据
        switch (this.pageType) {
          case 'tcp-client':
            response = await GetTCPClientData(item.rawData.id)
            if (response.success && response.data) {
              // 更新完整的客户端数据
              item.rawData = response.data
              this.selectedItem = item
              this.$emit('item-selected', item.rawData)
            } else {
              window.runtime.LogError('获取客户端数据失败: ' + (response.message || '未知错误'))
              // 即使获取详细数据失败，也使用现有数据进行切换
              this.selectedItem = item
              this.$emit('item-selected', item.rawData)
            }
            break
          case 'tcp-server':
            response = await GetTCPServerData(item.rawData.id)
            if (response.success && response.data) {
              item.rawData = response.data
              this.selectedItem = item
              this.$emit('item-selected', item.rawData)
            }
            break
          case 'udp-client':
            response = await GetUdpClientData(item.rawData.id)
            if (response.success && response.data) {
              item.rawData = response.data
              this.selectedItem = item
              this.$emit('item-selected', item.rawData)
            }
            break
          case 'udp-server':
            response = await GetUdpServerData(item.rawData.id)
            if (response.success && response.data) {
              item.rawData = response.data
              this.selectedItem = item
              this.$emit('item-selected', item.rawData)
            }
            break
          // ... 其他类型的处理保持不变
        }
      } catch (error) {
        window.runtime.LogError('获取数据失败: ' + error)
        // 发生错误时仍然切换选中项
        this.selectedItem = item
        this.$emit('item-selected', item.rawData)
      }
    },

    showAddModal() {
      if (this.pageType == 'tcp-client') {
        this.isModalVisibleTcpClient = true;
      } else if (this.pageType == 'tcp-server') {
        this.isModalVisibleTcpServer = true;
      } else if (this.pageType == 'udp-client') {
        this.isModalVisibleUdpClient = true;
      } else if (this.pageType == 'udp-server') {
        this.isModalVisibleUdpServer = true;
      }
    },

    closeAddModal() {
      if (this.pageType == 'tcp-client') {
        this.isModalVisibleTcpClient = false;
      } else if (this.pageType == 'tcp-server') {
        this.isModalVisibleTcpServer = false;
      } else if (this.pageType == 'udp-client') {
        this.isModalVisibleUdpClient = false;
      } else if (this.pageType == 'udp-server') {
        this.isModalVisibleUdpServer = false;
      }
      this.loadData(this.pageType); // 关闭模态框时刷新列表
    }
  }
}
</script>

<style scoped>
.list-container {
  display: flex;
  height: 100%;
  background-color: #f5f5f5;
}

.list-section {
  width: 250px;
  background-color: #ffffff;
  border-right: 1px solid #e8e8e8;
  display: flex;
  flex-direction: column;
}

.list-header {
  padding: 16px;
  border-bottom: 1px solid #e8e8e8;
  text-align: center;
}

.list-header h3 {
  margin: 0 0 16px 0;
  color: #333;
}

.button-container {
  display: flex;
  justify-content: center;
  width: 100%;
}

.button-group {
  display: flex;
  gap: 8px;
  justify-content: center;
}

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
}

.btn i {
  font-size: 12px;
}

.btn-add {
  background-color: #1890ff;
  color: white;
}

.btn-add:hover {
  background-color: #40a9ff;
}

.btn-delete {
  background-color: #ff4d4f;
  color: white;
}

.btn-delete:hover {
  background-color: #ff7875;
}

.btn-delete:disabled {
  background-color: #d9d9d9;
  cursor: not-allowed;
}

.list-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.list-item {
  padding: 12px 16px;
  margin: 4px 0;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.3s;
}

.list-item:hover {
  background-color: #f0f0f0;
}

.list-item.selected {
  background-color: #e6f7ff;
  border-right: 3px solid #1890ff;
}

.content-section {
  flex-grow: 1;
  padding: 20px;
  background-color: #ffffff;
}

.content-placeholder {
  color: #999;
  text-align: center;
  margin-top: 40px;
}

.content-detail {
  padding: 20px;
}

/* 滚动样式 */
.list-content::-webkit-scrollbar {
  width: 6px;
}

.list-content::-webkit-scrollbar-track {
  background: #f0f0f0;
}

.list-content::-webkit-scrollbar-thumb {
  background: #ccc;
  border-radius: 3px;
}

.list-content::-webkit-scrollbar-thumb:hover {
  background: #999;
}
</style> 