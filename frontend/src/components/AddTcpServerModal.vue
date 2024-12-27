<template>
  <div v-if="visible" class="modal-overlay">
    <div class="modal-content">
      <h2>{{ editMode ? '编辑配置' : '添加配置' }}</h2>
      <label>
        <span class="label-text">备注：</span>
        <input v-model="form.note" type="text" placeholder="请输入备注" />
      </label>
      <label>
        <span class="label-text">类型：</span>
        <input type="text" value="TCP 服务端" readonly />
      </label>
      <label>
        <span class="label-text">主机地址：</span>
        <input v-model="form.host" type="text" placeholder="请输入主机地址" />
      </label>
      <label>
        <span class="label-text">端口：</span>
        <input v-model="form.port" type="number" placeholder="请输入端口" />
      </label>
      <div class="modal-actions">
        <button class="btn-save" @click="save">保存</button>
        <button class="btn-cancel" @click="close">取消</button>
      </div>
    </div>
  </div>
</template>

<script>

import { AddTCPServer, UpdateTCPServer } from "../../wailsjs/go/control/FuncTcpServer";
import { ShowWarningDialog } from '../../wailsjs/go/main/App'

export default {
  name: 'AddTcpServerModal',
  props: {
    visible: {
      type: Boolean,
      required: true
    },
    onClose: {
      type: Function,
      required: true
    },
    editMode: {
      type: Boolean,
      required: true
    },
    initialData: {
      type: Object,
      default: () => ({})
    }
  },
  
  data() {
    return {
      form: {
        note: '',
        type: 'tcp',
        host: '',
        port: '',
      }
    }
  },
  // 监听visible的变化
  watch: {
    visible(newVal) {
      if (newVal && this.editMode && this.initialData) {
        // 当模态框显示且为编辑模式时，初始化表单数据
        this.initializeFormData()
      } else if (newVal && !this.editMode) {
        // 当为新增模式时，重置表单
        console.log('新增模式')
      }
    }
  },

  mounted() {
    if (this.editMode && this.initialData) {
      this.initializeFormData();
    }
  },

  methods: {
    initializeFormData() {
      this.form = {
        note: this.initialData.remark || '',
        type: 'tcp',
        host: this.initialData.host || '',
        port: this.initialData.port?.toString() || '',
      }
    },

    resetForm() {
      // 重置表单为初始状态
      this.form = {
        note: '',
        type: 'tcp',
        host: '',
        port: '',
      }
    },

    async save() {
      try {
        if (!this.form.host || !this.form.port) {
          window.runtime.LogInfo('主机和端口为必填项')
          return
        }

        const port = parseInt(this.form.port)
        if (isNaN(port) || port < 1 || port > 65535) {
          window.runtime.LogInfo('请输入有效的端口号(1-65535)')
          return
        }
        const tcpServer = {
          remark: this.form.note,
          type: this.form.type,
          host: this.form.host,
          port: port,
          status: this.editMode ? this.initialData.status : 'stopped',  // 保持原有状态
        }

        if (this.editMode) {
          tcpServer.id = this.initialData.id  // 确保包含ID
          const res = await UpdateTCPServer(tcpServer)
          if (res.success) {
            window.runtime.LogInfo('更新成功')
          } else {
            window.runtime.LogError('更新失败: ' + res.message)
            await ShowWarningDialog('错误', '更新失败: ' + res.message)
          }
        } else {
          const res = await AddTCPServer(tcpServer)
          if (res.success) {
            window.runtime.LogInfo('保存成功')
          } else {
            window.runtime.LogError('保存失败: ' + res.message)
            await ShowWarningDialog('错误', '保存失败: ' + res.message)
          }
        }
        
        this.close()
        // 触发刷新列表的事件
        this.$emit('refresh-list')
      } catch(err) {
        window.runtime.LogError(this.editMode ? '更新出错: ' : '保存出错: ' + err)
      }
    },

    close() {
      this.resetForm()  // 关闭时重置表单
      this.onClose()
    }
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
}

.modal-content {
  background: #ffffff;
  padding: 20px;
  border-radius: 8px;
  width: 500px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

h2 {
  color: #333;
  margin-bottom: 20px;
}

.label-text {
  display: inline-block;
  width: 120px; /* 设置统一的宽度 */
  margin-right: 10px; /* 标签文字与输入框之间的间距 */
  text-align: left; /* 文本左对齐 */
}

label {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
  color: black;
}

.uniform-select,
input[type="text"],
input[type="number"] {
  width: 100%;
  padding: 12px;
  margin-top: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: 14px;
  transition: border-color 0.3s;
  box-sizing: border-box;
}

.uniform-select:focus,
input[type="text"]:focus,
input[type="number"]:focus {
  border-color: #1890ff;
  outline: none;
}

.uniform-select:hover,
input[type="text"]:hover,
input[type="number"]:hover {
  border-color: #40a9ff;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.btn-save, .btn-cancel {
  padding: 10px 15px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
  margin-left: 10px;
}

.btn-save {
  background-color: #1890ff;
  color: white;
}

.btn-save:hover {
  background-color: #40a9ff;
}

.btn-cancel {
  background-color: #ff4d4f;
  color: white;
}

.btn-cancel:hover {
  background-color: #ff7875;
}

.repeat-send {
  display: flex;
  align-items: center;
  margin-top: 30px; /* 可根据需要��整间距 */
  justify-content: space-between; /* 使内容分散对齐 */
}

.repeat-send label {
  margin-right: 10px; /* 可根据需要调整间距 */
}

.repeat-send select,
.repeat-send input[type="text"] {
  margin-left: 10px; /* 可根据需要调整间距 */
  flex: 1; /* 使输入框和选择框占据剩余空间 */
}

.send-interval-label {
  margin-left: auto;
}

.send-interval-input {
  margin-left: 150px; /* 可根据需要调整间距 */
  width: auto; /* 使输入框宽度自适应 */
}
</style>