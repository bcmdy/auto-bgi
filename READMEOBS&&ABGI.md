# 配置OBS搭配ABGI的图文教程

## 配置OBS

下载地址：[OBS](https://obsproject.com/zh-cn)

下载安装好，就可以进行图文一步走系列，完全为图文教学

初始化
![OBS-1](./assets/OBS-1.png)
![OBS-2](./assets/OBS-2.png)
![OBS-3](./assets/OBS-3.png)

设置录制的窗口
![OBS添加录制窗口](assets/OBS添加录制窗口.png)
![OBS添加录制窗口1](assets/OBS添加录制窗口1.png)
![OBS添加录制窗口2](assets/OBS添加录制窗口2.png)

打开设置，找到输出，修改相关内容，点击确定
![OBS启动回放](assets/OBS启动回放.png)
点击【**启动回放缓存**】并保存一个视频
![OBSOBS保存回放](assets/OBS保存回放.png)

开启WebSocket服务器并复制相关信息
![OBS开启WebSocket](assets/OBS开启WebSocket.png)
![OBSWeb](assets/OBSWeb.png)

## 配置ABGI

请运行你autobgi程序，然后在浏览器输入你管理页面

找到配置文件，进入并填写相关内容
![abgi-1配置文件](assets/abgi-1配置文件.png)
![abgi-6关键字](assets/abgi-6关键字.png)
![abgi-2开启OBS回放缓冲](assets/abgi-2开启OBS回放缓冲.png)

**注意** 记得填写好后，在下面点击**保存**
![abgi-3填写相关信息](assets/abgi-3填写相关信息.png)

运行autobgi,浏览器打开管理页面，进入录屏管理
![abgi-4录屏管理](assets/abgi-4录屏管理.png)
![abgi-5手动录制](assets/abgi-5手动录制.png)
