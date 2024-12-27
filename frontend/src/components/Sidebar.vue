<template>
  <div class="sidebar">
    <div class="logo-section">
      <h2>网络调试工具</h2>
    </div>
    
    <div class="menu-section">
      <div class="menu-group">
        <div class="group-title">TCP</div>
        <div 
          class="menu-item" 
          :class="{ active: currentPage === 'tcp-client' }"
          @click="switchPage('tcp-client')"
        >
          <i class="fas fa-plug"></i>
          TCP客户端
        </div>
        <div 
          class="menu-item"
          :class="{ active: currentPage === 'tcp-server' }"
          @click="switchPage('tcp-server')"
        >
          <i class="fas fa-server"></i>
          TCP服务器
        </div>
      </div>

      <div class="menu-group">
        <div class="group-title">UDP</div>
        <div 
          class="menu-item"
          :class="{ active: currentPage === 'udp-client' }"
          @click="switchPage('udp-client')"
        >
          <i class="fas fa-plug"></i>
          UDP客户端
        </div>
        <div 
          class="menu-item"
          :class="{ active: currentPage === 'udp-server' }"
          @click="switchPage('udp-server')"
        >
          <i class="fas fa-server"></i>
          UDP服务器
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Sidebar',
  data() {
    return {
      currentPage: 'tcp-client'
    }
  },
  mounted() {
    // 监听来自菜单栏的页面切换事件
    window.runtime.EventsOn("switch-page", (page) => {
      this.currentPage = page;
    });
  },
  methods: {
    switchPage(page) {
      this.currentPage = page;
      this.$emit('page-changed', page);
    }
  }
}
</script>

<style scoped>
.sidebar {
  width: 250px;
  background-color: #1a1a1a;
  height: 100vh;
  color: #ffffff;
  display: flex;
  flex-direction: column;
  user-select: none;
}

.logo-section {
  padding: 20px;
  text-align: center;
  border-bottom: 1px solid #333;
}

.logo-section h2 {
  margin: 0;
  font-size: 1.2em;
  color: #fff;
}

.menu-section {
  flex: 1;
  padding: 20px 0;
  overflow-y: auto;
}

.menu-group {
  flex: 1;
  min-width: 0;
  /* background-color: #2d2d2d; */
  border-radius: 8px;
  padding: 15px;
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 20px;
}

.group-title {
  color: #666;
  font-size: 0.9em;
  margin-bottom: 15px;
  text-align: center;
  text-transform: uppercase;
  font-weight: bold;
  width: 100%;
}

.menu-item {
  padding: 12px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #999;
  text-align: center;
  border-radius: 6px;
  width: 100%;
}

.menu-item i {
  font-size: 1.2em;
  margin-bottom: 8px;
  display: block;
  text-align: center;
}

.menu-item div {
  font-size: 0.9em;
  width: 100%;
  text-align: center;
}

.menu-item:hover {
  background-color: #2d2d2d;
  color: #fff;
}

.menu-item.active {
  background-color: #363636;
  color: #fff;
  border-left: 3px solid #1890ff;
}

.footer-section {
  border-top: 1px solid #333;
  padding: 10px 0;
}

.version {
  text-align: center;
  font-size: 0.8em;
  color: #666;
  padding: 10px 0;
}

/* 滚动条样式 */
.menu-section::-webkit-scrollbar {
  width: 6px;
}

.menu-section::-webkit-scrollbar-track {
  background: #1a1a1a;
}

.menu-section::-webkit-scrollbar-thumb {
  background: #333;
  border-radius: 3px;
}

.menu-section::-webkit-scrollbar-thumb:hover {
  background: #444;
}
</style> 