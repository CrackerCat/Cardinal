package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/thanhpk/randstr"
)

type Team struct {
	gorm.Model

	Name     string
	Password string `json:"-"`
	Logo     string
	Score    int64
}

func (s *Service) GetAllTeams() (int, interface{}) {
	var teams []Team
	s.Mysql.Model(&Team{}).Find(&teams)
	return s.makeSuccessJSON(teams)
}

func (s *Service) NewTeams(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		Name string `binding:"required"`
		Logo string `binding:"required"`
	}
	var inputForm []InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	// 检查是否重复
	for _, item := range inputForm {
		var count int
		s.Mysql.Model(Team{}).Where(&Team{Name: item.Name}).Count(&count)
		if count != 0 {
			return s.makeErrJSON(400, 40001, "存在重复添加数据")
		}
	}

	type resultItem struct {
		Name     string
		Password string
	}
	var resultData []resultItem

	tx := s.Mysql.Begin()
	for _, item := range inputForm {
		password := randstr.String(16)
		newTeam := &Team{
			Name:     item.Name,
			Password: s.addSalt(password),
			Logo:     item.Logo,
		}
		if tx.Create(newTeam).RowsAffected != 1 {
			tx.Rollback()
			return s.makeErrJSON(500, 50000, "添加 Team 失败！")
		}
		resultData = append(resultData, resultItem{
			Name:     item.Name,
			Password: password,
		})
	}
	tx.Commit()
	return s.makeSuccessJSON(resultData)
}

func (s *Service) EditTeam(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID   uint   `binding:"required"`
		Name string `binding:"required"`
		Logo string `binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	// 检查 Team 是否存在
	var count int
	s.Mysql.Model(Team{}).Where(&Team{Model: gorm.Model{ID: inputForm.ID}}).Count(&count)
	if count == 0 {
		return s.makeErrJSON(404, 40400, "Team 不存在")
	}

	// 检查 Team Name 是否重复
	var repeatCheckTeam Team
	s.Mysql.Model(Team{}).Where(&Team{Name: inputForm.Name}).Find(&repeatCheckTeam)
	if repeatCheckTeam.Name != "" && repeatCheckTeam.ID != inputForm.ID {
		return s.makeErrJSON(400, 40001, "Team 重复")
	}

	newTeam := &Team{
		Name: inputForm.Name,
		Logo: inputForm.Logo,
	}
	tx := s.Mysql.Begin()
	if tx.Model(&Team{}).Where(&Team{Model: gorm.Model{ID: inputForm.ID}}).Updates(&newTeam).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50001, "修改 Team 失败！")
	}
	tx.Commit()

	return s.makeSuccessJSON("修改 Team 成功！")
}

func (s *Service) ResetTeamPassword(c *gin.Context) (int, interface{}) {
	type InputForm struct {
		ID uint `binding:"required"`
	}
	var inputForm InputForm
	err := c.BindJSON(&inputForm)
	if err != nil {
		return s.makeErrJSON(400, 40000, "Error payload")
	}

	// 检查 Team 是否存在
	var count int
	s.Mysql.Model(Team{}).Where(&Team{Model: gorm.Model{ID: inputForm.ID}}).Count(&count)
	if count == 0 {
		return s.makeErrJSON(404, 40400, "Team 不存在")
	}

	newPassword := randstr.String(16)
	tx := s.Mysql.Begin()
	if tx.Model(&Team{}).Where(&Team{Model: gorm.Model{ID: inputForm.ID}}).Updates(&Team{Password: s.addSalt(newPassword)}).RowsAffected != 1 {
		tx.Rollback()
		return s.makeErrJSON(500, 50001, "重置密码失败！")
	}
	tx.Commit()

	return s.makeSuccessJSON(newPassword)
}