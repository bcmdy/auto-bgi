package CDAwareAutoGather

import (
	"auto-bgi/ScriptGroup"
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/tools"
	"bufio"
	"encoding/json"
	"fmt"
	otiai10Copy "github.com/otiai10/copy"
	"github.com/tidwall/sjson"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UidInfo struct {
	UID               string              // 账户id
	CDAwareAutoGather []CDAwareAutoGather // 文件详情
}

type CDAwareAutoGather struct {
	TextName string   // txt名称
	Detail   []Detail // 文件详情
}

type Detail struct {
	FileName  string // 文件名
	CDTime    string // CD时间
	CDExpired bool   // CD是否到期
}

// ReadInfo 读取采集cd数据
// ReadInfo 函数用于读取并整理记录信息
// 返回一个包含所有记录的 UidInfo 切片
func (u *UidInfo) ReadInfo(status string) []UidInfo {
	// 构建基础目录路径，指向记录文件所在目录
	baseDir := filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript", "CD-Aware-AutoGather", "record")
	var allRecords []UidInfo

	// 遍历 record 目录下的所有子文件夹（账号）
	subDirs, err := os.ReadDir(baseDir)
	if err != nil {
		//panic(err)
		autoLog.Sugar.Errorf("读取目录失败:%s", err)
		return nil
	}

	for _, sub := range subDirs {
		if !sub.IsDir() {
			continue
		}
		uid := sub.Name()

		var uidInfo UidInfo
		uidInfo.UID = uid
		uidInfo.CDAwareAutoGather = []CDAwareAutoGather{}

		accountDir := filepath.Join(baseDir, uid)

		// 遍历该账号文件夹下的所有 txt 文件
		err := filepath.Walk(accountDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".txt") {
				record := readAndFormatTxt(path, info.Name(), status)
				if record != nil {
					uidInfo.CDAwareAutoGather = append(uidInfo.CDAwareAutoGather, *record)
				}
			}
			return nil
		})
		if err != nil {

			autoLog.Sugar.Errorf("CD遍历账号文件夹出错:%s,%d", uid, err)
		}

		allRecords = append(allRecords, uidInfo)
	}

	// 输出查看
	return allRecords
}

func readAndFormatTxt(filePath string, txtName, status string) *CDAwareAutoGather {
	file, err := os.Open(filePath)
	if err != nil {
		autoLog.Sugar.Errorf("打开文件失败:%s", err)
		return nil
	}
	defer file.Close()

	cda := CDAwareAutoGather{
		TextName: txtName,
		Detail:   []Detail{},
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		filename := parts[0]
		timeStr := parts[len(parts)-1]

		t, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			autoLog.Sugar.Errorf("解析时间失败:%v", err)
			continue
		}

		detail := Detail{
			FileName:  filename,
			CDTime:    t.Format("2006-01-02 15:04:05"),
			CDExpired: t.Before(time.Now()),
		}

		// 根据 status 筛选
		if (status == "1" && detail.CDExpired) ||
			(status == "2" && !detail.CDExpired) ||
			(status == "3") {
			cda.Detail = append(cda.Detail, detail)
		}
	}

	if err := scanner.Err(); err != nil {
		autoLog.Sugar.Errorf("读取文件出错:%v", err)
	}

	return &cda
}

// 提取出所有CD采集的材料，加入背包统计
func (u *UidInfo) CDAllMaterial() []string {

	material := make(map[string]int)

	materialConstant := abgiConstant.Material

	readInfo := u.ReadInfo("3")
	for _, info := range readInfo {
		for _, gather := range info.CDAwareAutoGather {
			//获取材料名称
			for _, s := range materialConstant {
				if strings.Contains(gather.TextName, s) {
					material[s] = 1
				}
			}
		}
	}

	statistics := strings.Split(config.Cfg.BagStatistics, ",")
	for _, s := range statistics {
		material[s] = 1
	}
	var res []string
	for k, _ := range material {

		autoLog.Sugar.Infof("获取到材料：%s", k)

		res = append(res, k)
	}

	//加入到背包统计
	config.Cfg.BagStatistics = strings.Join(res, ",")
	err := config.WriteConfig()
	if err != nil {
		autoLog.Sugar.Errorf("背包统计加入失败:%s", err)
	}

	return res
}

type settings struct {
	Name string
}

// 读取CDAwareAutoGather所有的路线
func (u *UidInfo) ReadAllRoute(CDAwareAutoGatherGroup string) map[string]interface{} {

	//查询配置组的勾选路线
	CDAwareAutoGatherGroupFilename := config.Cfg.BetterGIAddress + "\\User\\ScriptGroup\\" + CDAwareAutoGatherGroup + ".json"
	// 读取 狗粮联机JSON
	CDGroupData, err := os.ReadFile(CDAwareAutoGatherGroupFilename)
	if err != nil {
		autoLog.Sugar.Errorf("CD管理的自动采集[%s]失败:%d", CDAwareAutoGatherGroup, err)
	}
	var scriptGroupConfig ScriptGroup.ScriptGroupConfig
	err = json.Unmarshal(CDGroupData, &scriptGroupConfig)
	if err != nil {
		autoLog.Sugar.Errorf("CDAwareAutoGatherGroup配置组失败失败:%d", err)
	}
	JsScriptSettingsObject := make(map[string]interface{})
	for _, project := range scriptGroupConfig.Projects {
		if project.FolderName == "CD-Aware-AutoGather" {
			JsScriptSettingsObject = project.JsScriptSettingsObject
		}

	}

	//查询配置文件settings
	filename := config.Cfg.BetterGIAddress + "\\User\\JsScript\\CD-Aware-AutoGather\\settings.json"
	// 读取 狗粮联机JSON
	RouteData, err := os.ReadFile(filename)
	if err != nil {
		autoLog.Sugar.Errorf("读取CDAwareAutoGather所有的路线[CD-Aware-AutoGather]失败:%d", err)
	}
	var settings []settings
	err = json.Unmarshal(RouteData, &settings)
	if err != nil {
		autoLog.Sugar.Errorf("读取CDAwareAutoGather所有的路线[CD-Aware-AutoGather]失败:%d", err)
	}

	cdAwareAutoGatherRoute := make(map[string]interface{})
	for _, setting := range settings {
		if strings.HasPrefix(setting.Name, "OPT_") {
			if JsScriptSettingsObject[setting.Name] != nil {
				cdAwareAutoGatherRoute[setting.Name] = JsScriptSettingsObject[setting.Name]
			} else {
				cdAwareAutoGatherRoute[setting.Name] = true
			}

		}
	}

	u.WriteAllRoute(CDAwareAutoGatherGroup, cdAwareAutoGatherRoute)

	return cdAwareAutoGatherRoute

}

// 覆盖采集路线
func (u *UidInfo) WriteAllRoute(CDAwareAutoGatherGroup string, cdAwareAutoGatherRoutes map[string]interface{}) {
	//查询配置组的勾选路线
	CDAwareAutoGatherGroupFilename := config.Cfg.BetterGIAddress + "\\User\\ScriptGroup\\" + CDAwareAutoGatherGroup + ".json"
	// 读取 狗粮联机JSON
	CDGroupData, err := os.ReadFile(CDAwareAutoGatherGroupFilename)
	if err != nil {
		autoLog.Sugar.Errorf("CD管理的自动采集[%s]失败:%d", CDAwareAutoGatherGroup, err)
	}
	var scriptGroupConfig ScriptGroup.ScriptGroupConfig
	err = json.Unmarshal(CDGroupData, &scriptGroupConfig)
	if err != nil {
		autoLog.Sugar.Errorf("CDAwareAutoGatherGroup配置组失败失败:%d", err)
	}
	newData := CDGroupData
	for i, project := range scriptGroupConfig.Projects {
		if project.FolderName == "CD-Aware-AutoGather" {
			jsScriptSettingsObject := fmt.Sprintf("projects.%d.jsScriptSettingsObject", i)
			newData, err = sjson.SetBytes(newData, jsScriptSettingsObject, cdAwareAutoGatherRoutes)
			break
		}
	}

	// 写回文件
	if err := os.WriteFile(CDAwareAutoGatherGroupFilename, newData, 0644); err != nil {

		autoLog.Sugar.Errorf("写入 CD管理的自动采集[%s]失败:%d", CDAwareAutoGatherGroupFilename, err)
	}
}

var CKRouter = map[string]string{
	"矿物\\水晶块\\诺艾尔的提瓦特矿闻录\\火山\\层岩巨渊\\地下矿区": "矿物\\水晶块\\诺艾尔的提瓦特矿闻录@火山\\层岩巨渊·地下矿区",
	"矿物\\紫晶块\\紫晶块\\大剑\\\\蜜柑魚":             "矿物\\紫晶块\\紫晶块[大剑]@蜜柑魚",
	"食材与炼金\\日落果\\有草神\\今天下雨w":              "食材与炼金\\日落果\\有草神@今天下雨w",
	"食材与炼金\\松茸\\有草神\\未知作者":                "食材与炼金\\松茸\\有草神@未知作者",
	"矿物\\萃凝晶\\萃凝晶\\火神赶路\\大剑\\无限不循环":       "矿物\\萃凝晶\\萃凝晶-火神赶路-大剑@无限不循环",
	"食材与炼金\\枣椰\\有草神\\不瘦五十斤不改名":            "食材与炼金\\枣椰\\有草神@不瘦五十斤不改名",
	"食材与炼金\\金鱼草\\金鱼草\\白白喵":                "食材与炼金\\金鱼草\\金鱼草@白白喵",
}

// 更新所有CD管理
func (u *UidInfo) UpdateAllCD() {
	info := u.ReadInfo("3")
	for _, router := range info {
		for _, detail := range router.CDAwareAutoGather {
			TextName := strings.ReplaceAll(detail.TextName, "_", "\\")
			TextName = strings.ReplaceAll(TextName, ".txt", "")
			//找到本地仓库文件夹
			localJoin := filepath.Join(config.Cfg.BetterGIAddress, "User", "AutoPathing", TextName)
			ckJoin := filepath.Join(config.Cfg.BetterGIAddress, "Repos", "bettergi-scripts-list-git", "repo", "pathing", TextName)

			if _, err := os.Stat(localJoin); os.IsNotExist(err) {
				cKRouter := CKRouter[TextName]
				if cKRouter != "" {
					TextName = cKRouter
					localJoin = filepath.Join(config.Cfg.BetterGIAddress, "User", "AutoPathing", cKRouter)
					ckJoin = filepath.Join(config.Cfg.BetterGIAddress, "Repos", "bettergi-scripts-list-git", "repo", "pathing", cKRouter)
				} else {
					autoLog.Sugar.Error("=============", TextName)
					autoLog.Sugar.Errorf("本地仓库文件夹不存在:%d", err)
					continue
				}
			}
			dirs, err := tools.CompareDirs(localJoin, ckJoin)
			if err != nil {
				autoLog.Sugar.Errorf("比较文件夹失败:%d", err)
			}
			if !dirs {
				autoLog.Sugar.Infof("文件有更新:%s", TextName)
				err = otiai10Copy.Copy(ckJoin, localJoin)
				if err != nil {
					autoLog.Sugar.Errorf("更新文件失败:%d", err)
				} else {
					autoLog.Sugar.Infof("更新文件成功:%s", TextName)
				}
			}

		}
	}
}
