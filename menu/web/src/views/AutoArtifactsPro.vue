<template>
  <div class="container">
    <!-- 顶部标题栏 -->
    <header class="topbar">
      <div class="title-row">
        <h1 class="main-title">
          <span class="title-glow">✦</span>
          {{ title }}
        </h1>
        <span v-if="jsVersion" class="title-badge">JS {{ jsVersion }}</span>
      </div>

      <div class="button-group">
        <button @click="goToHome" class="btn btn-ghost">
          <span class="btn-icon">⟲</span>
          返回主页
        </button>
      </div>
    </header>

    <!-- 主面板 -->
    <section class="panel" id="executionPanel">
      <!-- 状态：加载中 -->
      <div v-if="loading" class="state state-loading">
        <div class="loading-spinner"></div>
        <p class="state-text">正在召唤数据中…</p>
        <p class="state-sub">（请稍候，命运齿轮正在转动）</p>
      </div>

      <!-- 状态：无数据 -->
      <div v-else-if="items.length === 0" class="state state-empty">
        <div class="empty-icon">☁</div>
        <p class="state-text">暂无数据</p>
        <p class="state-sub">也许后端还没把“世界线”同步过来。</p>
      </div>

      <!-- 数据列表 -->
      <div v-else class="grid">
        <article v-for="(item, index) in items" :key="index" class="info-box">
          <!-- 角落轮播装饰 -->
          <div
            v-if="images.length > 0"
            class="swiper corner-icon-swiper"
            :ref="`swiper-${index}`"
            aria-label="corner swiper"
          >
            <div class="swiper-wrapper">
              <div v-for="(image, imgIndex) in images" :key="imgIndex" class="swiper-slide">
                <img :src="`/static/image/${image}`" :alt="`image-${imgIndex}`" />
              </div>
            </div>
          </div>

          <!-- 文件头 -->
          <div class="file-header">
            <h3 class="file-title">
              <span class="file-icon">📄</span>
              <span class="file-name" :title="item.FileName">{{ item.FileName }}</span>
            </h3>

            <button class="btn btn-skill" @click="getAutoArtifactsPro2Btn(item.FileName)">
              <span class="btn-icon">📈</span>
              转为折线图
              <span class="btn-spark">✧</span>
            </button>
          </div>

          <!-- meta -->
          <div class="meta">
            <div class="meta-row">
              <span class="meta-tag">MARK</span>
              <span class="meta-value">{{ item.Mark }}</span>
            </div>
            <div class="meta-row">
              <span class="meta-tag">TIME</span>
              <span class="meta-value">{{ item.ActivateTime }}</span>
            </div>
          </div>

          <!-- detail -->
          <div class="file-details">
            <div class="detail-head">
              <span class="detail-title">✦ Detail Logs</span>
              <span class="detail-hint">（逐条展开）</span>
            </div>

            <div class="detail-item" v-for="(detail, detailIndex) in item.Detail" :key="detailIndex">
              <span class="detail-dot">◆</span>
              <span class="detail-text">{{ detail }}</span>
            </div>
          </div>
        </article>
      </div>
    </section>

    <!-- 底部装饰 -->
    <footer class="footer">
      <span class="footer-line">「愿收益与你同在。」</span>
    </footer>
  </div>
</template>

<script>
import { ref, onMounted, nextTick, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { Swiper } from 'swiper'
import 'swiper/css'
import 'swiper/css/effect-fade'
import { apiMethods } from '@/utils/api'

export default {
  name: 'AutoArtifactsPro',
  setup() {
    const router = useRouter()

    // 响应式数据
    const title = ref('狗粮批发-联机收益查询')
    const jsVersion = ref('')
    const items = ref([])
    const images = ref([])
    const loading = ref(false)

    // ✅ 仅用于销毁 swiper，避免离开页面资源残留（不改业务逻辑）
    const swipers = ref([])

    // 方法
    const goToHome = () => {
      router.push('/')
    }

    const getAutoArtifactsPro2Btn = (fileName) => {
      router.push({ path: '/getAutoArtifactsPro2', query: { fileName } })
    }

    const loadImages = async () => {
      try {
        const data = await apiMethods.getImages()
        images.value = data.data || []
      } catch (err) {
        console.error('Failed to load images:', err)
        // fallback
        images.value = ['bd.jpg', 'ff.png', 'ng.jpg', 'sh.jpg']
      }
    }

    const initializeSwipers = async () => {
      await nextTick()

      // 先清理旧的（例如热更新/重复进入）
      swipers.value.forEach((s) => {
        try {
          s && s.destroy && s.destroy(true, true)
        } catch (e) {}
      })
      swipers.value = []

      const swiperContainers = document.querySelectorAll('.corner-icon-swiper')
      swiperContainers.forEach((swiperContainer) => {
        if (images.value.length > 0) {
          const ins = new Swiper(swiperContainer, {
            loop: true,
            autoplay: {
              delay: 5000,
              disableOnInteraction: false
            },
            effect: 'fade',
            fadeEffect: {
              crossFade: true
            }
          })
          swipers.value.push(ins)
        }
      })
    }

    const loadData = async () => {
      try {
        loading.value = true
        const data = await apiMethods.getAutoArtifactsPro()

        if (data) {
          title.value = data.title || 'AutoArtifactsPro'
          jsVersion.value = data.JsVersion || ''
          items.value = data.items || []
        }
      } catch (err) {
        console.error('Failed to load AutoArtifactsPro data:', err)
        items.value = [
          {
            FileName: 'example1.txt',
            Detail: ['暂无数据，请检查后端服务'],
            Mark: '暂无数据',
            ActivateTime: ''
          }
        ]
      } finally {
        loading.value = false
      }
    }

    // 生命周期
    onMounted(async () => {
      await loadImages()
      await loadData()
      await initializeSwipers()
    })

    onBeforeUnmount(() => {
      // ✅ 清理 swiper，避免页面离开后继续占用
      swipers.value.forEach((s) => {
        try {
          s && s.destroy && s.destroy(true, true)
        } catch (e) {}
      })
      swipers.value = []
    })

    return {
      title,
      jsVersion,
      items,
      images,
      loading,
      goToHome,
      getAutoArtifactsPro2Btn
    }
  }
}
</script>

<style scoped>
:root {
  /* ✅ 更亮更饱和的二次元霓虹配色 */
  --bg0: #fff3fb;
  --bg1: #f2fbff;
  --card: rgba(255, 255, 255, 0.72);
  --card2: rgba(255, 255, 255, 0.56);

  --stroke: rgba(255, 105, 180, 0.35);
  --stroke2: rgba(120, 170, 255, 0.26);

  --text: #4a225a;
  --muted: rgba(74, 34, 90, 0.68);

  --pink: #ff4fc3;
  --violet: #7c5cff;
  --cyan: #00d9ff;
  --gold: #ffcc5c;

  --shadow: 0 18px 48px rgba(255, 79, 195, 0.18);
  --shadow2: 0 10px 28px rgba(124, 92, 255, 0.15);

  --radius-xl: 22px;
  --radius-lg: 18px;
  --radius-md: 14px;

  --maxw: 980px;
}

* { box-sizing: border-box; margin: 0; padding: 0; }

.container{
  min-height: 100vh;
  padding: 18px 16px 30px;
  color: var(--text);
  font-family: ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, "PingFang SC",
    "Microsoft YaHei", "Noto Sans CJK SC", sans-serif;

  /* ✅ 明亮的粉紫青渐变背景 */
  background:
    radial-gradient(1200px 800px at 15% 10%, rgba(124, 92, 255, 0.26), transparent 60%),
    radial-gradient(900px 600px at 85% 20%, rgba(0, 217, 255, 0.22), transparent 55%),
    radial-gradient(700px 500px at 50% 90%, rgba(255, 79, 195, 0.24), transparent 60%),
    linear-gradient(160deg, var(--bg0), var(--bg1));
  position: relative;
  overflow: hidden;
}

/* ✅ 更明显的点阵装饰（但不会压字） */
.container::before{
  content:"";
  position:absolute;
  inset:-40px;
  background-image: radial-gradient(rgba(255, 79, 195, 0.18) 1px, transparent 1px);
  background-size: 22px 22px;
  opacity: .45;
  pointer-events:none;
  animation: floatGrid 10s linear infinite;
}
@keyframes floatGrid { 0%{transform:translateY(0)} 100%{transform:translateY(22px)} }

/* 顶部栏 */
.topbar{
  max-width: var(--maxw);
  margin: 0 auto 14px;
  display:flex;
  align-items:flex-start;
  justify-content:space-between;
  gap:12px;
  position:relative;
  z-index:2;

  padding: 14px;
  border-radius: var(--radius-xl);
  background: linear-gradient(135deg, rgba(255,255,255,0.78), rgba(255,255,255,0.55));
  border: 1px solid var(--stroke);
  box-shadow: var(--shadow2);
  backdrop-filter: blur(12px);
}

/* 标题 */
.title-row{ display:flex; align-items:center; gap:10px; flex-wrap:wrap; }
.main-title{
  font-size: 1.55rem;
  font-weight: 900;
  letter-spacing: .5px;
  color: #4a225a;
  text-shadow:
    0 0 18px rgba(255,79,195,0.25),
    0 0 28px rgba(124,92,255,0.20);
  display:flex;
  align-items:center;
  gap:10px;
}

.title-glow{
  display:inline-block;
  color: var(--cyan);
  filter: drop-shadow(0 0 12px rgba(0,217,255,0.55));
  animation: pulseRune 1.6s ease-in-out infinite;
}
@keyframes pulseRune { 0%,100%{transform:translateY(0) scale(1);opacity:.85} 50%{transform:translateY(-1px) scale(1.08);opacity:1} }

.title-badge{
  font-size:.78rem;
  padding:6px 10px;
  border-radius:999px;
  border: 1px solid rgba(255,79,195,0.30);
  color:#6a2d86;
  background: linear-gradient(135deg, rgba(255,79,195,0.22), rgba(124,92,255,0.18));
  box-shadow: 0 0 0 1px rgba(255,255,255,0.4) inset;
  white-space:nowrap;
}

.button-group{ display:flex; justify-content:flex-end; align-items:center; gap:10px; flex-wrap:wrap; }

/* 按钮：更亮更霓虹 */
.btn{
  border: 1px solid rgba(255,79,195,0.35);
  border-radius: 999px;
  padding: 10px 14px;
  cursor: pointer;
  font-weight: 900;
  letter-spacing: .3px;
  color: #5a216f;
  background: rgba(255,255,255,0.72);
  box-shadow: 0 10px 26px rgba(255,79,195,0.18);
  transition: transform .18s ease, box-shadow .18s ease, border-color .18s ease, filter .18s ease;
  display:inline-flex;
  align-items:center;
  gap:10px;
  user-select:none;
}
.btn:hover{
  transform: translateY(-1px);
  border-color: rgba(0,217,255,0.55);
  box-shadow: 0 16px 36px rgba(0,217,255,0.18), 0 12px 30px rgba(255,79,195,0.15);
  filter: saturate(1.15);
}
.btn:active{ transform: translateY(1px) scale(.99); }

.btn-icon{
  width:22px;height:22px;
  display:inline-flex;
  align-items:center;justify-content:center;
  border-radius: 10px;
  background: rgba(0,217,255,0.10);
  border: 1px solid rgba(0,217,255,0.25);
}

.btn-ghost{ background: rgba(255,255,255,0.65); }

.btn-skill{
  border: 1px solid rgba(0,217,255,0.45);
  color:#4a225a;
  background: linear-gradient(135deg, rgba(0,217,255,0.18), rgba(255,79,195,0.16));
  position:relative;
  overflow:hidden;
}
.btn-skill::before{
  content:"";
  position:absolute;
  inset:-2px;
  background:
    radial-gradient(240px 120px at 20% 20%, rgba(0,217,255,0.30), transparent 55%),
    radial-gradient(220px 120px at 80% 60%, rgba(255,79,195,0.28), transparent 55%);
  opacity:.95;
  pointer-events:none;
}
.btn-spark{
  margin-left:2px;
  color: var(--gold);
  filter: drop-shadow(0 0 10px rgba(255,204,92,0.6));
  animation: twinkle 1.2s ease-in-out infinite;
}
@keyframes twinkle { 0%,100%{opacity:.6;transform:translateY(0)} 50%{opacity:1;transform:translateY(-1px)} }

/* 主面板：更亮 */
.panel{
  max-width: var(--maxw);
  margin: 0 auto;
  position: relative;
  z-index:2;
  border-radius: var(--radius-xl);
  background: linear-gradient(135deg, rgba(255,255,255,0.75), rgba(255,255,255,0.50));
  border: 1px solid rgba(255,79,195,0.22);
  box-shadow: var(--shadow);
  backdrop-filter: blur(14px);
  padding: 14px;
}

/* 状态 */
.state{
  padding: 40px 12px;
  text-align:center;
  border-radius: var(--radius-lg);
  border: 1px dashed rgba(255,79,195,0.25);
  background: rgba(255,255,255,0.55);
}
.state-text{ font-size:1.1rem; font-weight:900; margin-top:14px; color:#4a225a; }
.state-sub{ margin-top:8px; font-size:.92rem; color: var(--muted); }
.empty-icon{ font-size:42px; filter: drop-shadow(0 0 16px rgba(255,79,195,0.25)); }

.loading-spinner{
  width:44px;height:44px;margin:0 auto;
  border-radius:999px;
  border: 3px solid rgba(124,92,255,0.20);
  border-top-color: rgba(0,217,255,0.95);
  animation: spin 1s linear infinite;
  box-shadow: 0 0 22px rgba(0,217,255,0.22);
}
@keyframes spin { to{transform:rotate(360deg)} }

/* 列表 */
.grid{ display:grid; gap:14px; grid-template-columns: 1fr; }

.info-box{
  position:relative;
  overflow:hidden;
  border-radius: var(--radius-xl);
  border: 1px solid rgba(124,92,255,0.18);
  background: linear-gradient(135deg, rgba(255,255,255,0.78), rgba(255,255,255,0.55));
  box-shadow: 0 18px 44px rgba(124,92,255,0.14);
  padding: 16px;
}
.info-box::before{
  content:"";
  position:absolute;
  inset:-1px;
  border-radius: var(--radius-xl);
  background: linear-gradient(90deg, rgba(255,79,195,0.35), rgba(124,92,255,0.30), rgba(0,217,255,0.28));
  opacity: .65;
  filter: blur(14px);
  pointer-events:none;
}
.info-box:hover{ transform: translateY(-1px); transition: transform .18s ease; }

/* 文件头 */
.file-header{
  display:flex;
  align-items:center;
  justify-content:space-between;
  gap:12px;
  flex-wrap:wrap;
  padding-bottom:12px;
  margin-bottom:12px;
  border-bottom: 1px solid rgba(124,92,255,0.14);
  position:relative;
  z-index:2;
}
.file-title{
  display:flex;align-items:center;gap:10px;
  font-size:1.1rem;font-weight:900;
  color:#4a225a;
}
.file-icon{ filter: drop-shadow(0 0 10px rgba(255,79,195,0.25)); }
.file-name{
  max-width: 64vw;
  overflow:hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* meta */
.meta{
  display:grid;
  gap:10px;
  padding: 10px 12px;
  border-radius: var(--radius-lg);
  border: 1px solid rgba(255,79,195,0.16);
  background: rgba(255,255,255,0.55);
  position:relative;
  z-index:2;
}
.meta-row{ display:flex; align-items:baseline; justify-content:space-between; gap:12px; flex-wrap:wrap; }
.meta-tag{
  font-size:.78rem;
  letter-spacing:1px;
  font-weight:900;
  color:#6a2d86;
  padding: 4px 8px;
  border-radius:999px;
  border: 1px solid rgba(0,217,255,0.22);
  background: rgba(0,217,255,0.08);
}
.meta-value{
  color:#4a225a;
  font-weight:900;
  text-shadow: 0 0 14px rgba(255,79,195,0.10);
  word-break: break-word;
}

/* detail */
.file-details{
  margin-top:12px;
  border-radius: var(--radius-lg);
  border: 1px solid rgba(124,92,255,0.14);
  background: rgba(255,255,255,0.60);
  padding: 12px;
  position:relative;
  z-index:2;
}
.detail-head{ display:flex; align-items:baseline; justify-content:space-between; gap:12px; margin-bottom:10px; }
.detail-title{ font-weight:900; color:#4a225a; text-shadow: 0 0 18px rgba(0,217,255,0.12); }
.detail-hint{ font-size:.85rem; color: var(--muted); }

.detail-item{
  display:flex;
  align-items:flex-start;
  gap:10px;
  padding: 10px;
  border-radius: var(--radius-md);
  border: 1px solid rgba(255,79,195,0.12);
  background: rgba(255,255,255,0.68);
  margin-top: 8px;
  transition: transform .15s ease, filter .15s ease;
}
.detail-item:hover{ transform: translateX(2px); filter: saturate(1.1); }

.detail-dot{
  color: var(--cyan);
  filter: drop-shadow(0 0 10px rgba(0,217,255,0.35));
  line-height: 1.4;
}
.detail-text{
  flex:1;
  color:#4a225a;
  word-break: break-word;
  line-height:1.5;
}

/* 角落轮播装饰（更亮但不抢戏） */
.corner-icon-swiper{
  position:absolute;
  right: 10px;
  bottom: 10px;
  width: 200px;
  height: 340px;
  border-radius: 16px;
  overflow:hidden;

  z-index:1;
  pointer-events:none;
  opacity: 0.18;

  border: 1px solid rgba(0,217,255,0.25);
  box-shadow: 0 18px 44px rgba(0,217,255,0.10), 0 18px 44px rgba(255,79,195,0.10);
  backdrop-filter: blur(6px);
}
.corner-icon-swiper .swiper-slide img{ width:100%; height:100%; object-fit:cover; }

/* footer */
.footer{
  max-width: var(--maxw);
  margin: 14px auto 0;
  text-align:center;
  color: rgba(74,34,90,0.70);
  font-weight: 900;
  position:relative;
  z-index:2;
}
.footer-line{
  display:inline-block;
  padding:10px 14px;
  border-radius:999px;
  border: 1px solid rgba(255,79,195,0.18);
  background: rgba(255,255,255,0.62);
  backdrop-filter: blur(10px);
}

/* 响应式 */
@media (min-width: 768px){
  .main-title{ font-size: 1.85rem; }
  .panel{ padding: 16px; }
  .info-box{ padding: 18px; }
  .file-name{ max-width: 520px; }
  .corner-icon-swiper{ width: 220px; height: 380px; opacity: .16; }
}
@media (max-width: 520px){
  .topbar{ padding: 12px; }
  .main-title{ font-size: 1.18rem; }
  .btn{ padding: 9px 12px; font-size: .92rem; }
  .file-title{ font-size: 1rem; }
  .corner-icon-swiper{ width: 140px; height: 220px; right: 8px; bottom: 8px; opacity: .12; }
  .meta-row{ justify-content:flex-start; }
}
</style>
