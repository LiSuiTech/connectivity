<template>
  <div id="app">
    <Sidebar @page-changed="handlePageChange" />
    <div class="main-content">
      <ListView 
        v-if="showListView" 
        :pageType="currentPage"
        @item-selected="handleItemSelected"
        @add-item="handleAddItem"
      />
      <WelcomePage v-else />
    </div>
  </div>
</template>

<script>
import Sidebar from './components/Sidebar.vue';
import ListView from './components/ListView.vue';
import WelcomePage from './components/WelcomePage.vue';

export default {
  components: {
    Sidebar,
    ListView,
    WelcomePage
  },
  data() {
    return {
      currentPage: 'tcp-client'
    }
  },
  computed: {
    showListView() {
      // 定义哪些页面需要显示列表视图
      const listViewPages = [
        'tcp-client', 
        'tcp-server', 
        'udp-client', 
        'udp-server',
        'ping',
        'port-scanner'
      ];
      return listViewPages.includes(this.currentPage);
    }
  },
  methods: {
    handlePageChange(page) {
      this.currentPage = page;
    },
    handleItemSelected(item) {
      // 处理列表项选中事件
      console.log('Selected item:', item);
    },
    handleAddItem(type) {
      // 处理添加项目事件
      console.log('Add item for type:', type);
    }
  }
}
</script>

<style>
#app {
  display: flex;
}
.main-content {
  flex: 1;
}
</style>