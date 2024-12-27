<template>
  <div :style="{ width: windowWidth + 'px', height: windowHeight + 'px' }" class="welcome-page">
    <!-- 欢迎页面内容 -->
    <h1 class="welcome-title">欢迎使用网络工具🛜</h1>
  </div>
</template>

<script>
export default {
  data() {
    return {
      windowWidth: window.innerWidth - 250 - 250, // 当前窗口宽度减去列表和侧边栏宽度
      windowHeight: window.innerHeight, // 当前窗口高度
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
      this.windowWidth = window.innerWidth - 250 - 250; // 更新窗口宽度，减去列表和侧边栏宽度
      this.windowHeight = window.innerHeight; // 更新窗口高度
    }
  }
}
</script>

<style scoped>
.welcome-page {
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #ffffff;
}

.welcome-title {
  color: black; /* 设置字体颜色为黑色 */
}
</style> 