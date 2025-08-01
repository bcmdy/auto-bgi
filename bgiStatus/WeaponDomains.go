package bgiStatus

import (
	"auto-bgi/config"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type WeaponDomain struct {
	DomainName   string
	Weekday      int
	MaterialName string
}

func (domain *WeaponDomain) QueryWeaponDomainHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从查询参数读取
		domainName := c.Query("domain_name")
		weekdayStr := c.Query("weekday")
		materialName := c.Query("material_name")

		var weekday int
		if weekdayStr != "" {
			// 解析weekday
			// 忽略错误或你可以返回400
			fmt.Sscanf(weekdayStr, "%d", &weekday)
		}

		// 动态构建SQL和参数
		baseSQL := "SELECT domain_name, weekday, material_name FROM weapon_domains WHERE 1=1"
		var params []interface{}

		if domainName != "" {
			baseSQL += " AND domain_name = ?"
			params = append(params, domainName)
		}
		if weekday != 0 {
			baseSQL += " AND weekday = ?"
			params = append(params, weekday)
		}
		if materialName != "" {
			baseSQL += " AND material_name = ?"
			params = append(params, materialName)
		}

		rows, err := config.DB.Query(baseSQL, params...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败: " + err.Error()})
			return
		}
		defer rows.Close()

		var results []TalentDomain
		for rows.Next() {
			var td TalentDomain
			if err := rows.Scan(&td.DomainName, &td.Weekday, &td.MaterialName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "扫描数据失败: " + err.Error()})
				return
			}
			results = append(results, td)
		}

		c.JSON(http.StatusOK, results)
	}
}

// 查询所有武器升级材料副本
func (talentDomain *TalentDomain) QueryAllWeaponDomains() ([]string, error) {

	rows, err := config.DB.Query(`SELECT DISTINCT domain_name FROM weapon_domains ORDER BY domain_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		domains = append(domains, name)
	}
	return domains, nil
}

// 查询所有武器升级材料
func (domain *WeaponDomain) QueryAllWeaponMaterials() ([]string, error) {
	rows, err := config.DB.Query(`SELECT DISTINCT material_name FROM weapon_domains ORDER BY material_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var talents []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		talents = append(talents, name)
	}
	return talents, nil
}
