package routes

import (
	"github.com/arifzaenal/gin-full-api/config"
	"github.com/arifzaenal/gin-full-api/models"
	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"strconv"
	"time"
)

func GetHome(c *gin.Context) {

	items := []models.Article{}
	config.DB.Find(&items)

	c.JSON(200, gin.H{
		"status" : "berhasil",
		"message" : items,
	})
}

func GetArticle(c *gin.Context){
	slug := c.Param("slug")

	var item models.Article

	if config.DB.First(&item, "slug = ?", slug).RecordNotFound() {
		c.JSON(404, gin.H{"status": "error", "message" : "record not found"})
		c.Abort()
		return
	}

	c.JSON(200, gin.H{
		"status" : "berhasil",
		"message" : item,
	})
}

func PostArticle(c *gin.Context) {

	var oldItem models.Article
	slug := slug.Make(c.PostForm("title"))

	if !config.DB.First(&oldItem, "slug = ?", slug).RecordNotFound() {
		slug = slug + "-" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	item := models.Article {
		Title 	: c.PostForm("title"),
		Desc	: c.PostForm("desc"),
		Tag		: c.PostForm("tag"),
		Slug	: slug,
		UserID	: uint(c.MustGet("jwt_user_id").(float64)),
	}

	config.DB.Create(&item)

	c.JSON(200, gin.H{
		"status": 	"berhasil menambahkan artikel",
		"data"	:	item,
	})
}

func GetArticleByTag(c *gin.Context) {
	tag := c.Param("tag")

	var items []models.Article

	if config.DB.Where("tag LIKE ?", "%" + tag + "%").Find(&items).RecordNotFound() {
		c.JSON(404, gin.H{
			"status"  : "error",
			"message" : "record not found",
		})
		c.Abort()
		return
	}

	c.JSON(200, gin.H{
		"status"  : "berhasil",
		"message" : items,
	})
}

func UpdateArticle(c *gin.Context) {

	id := c.Param("id")
	var item models.Article

	if config.DB.First(&item, "id = ?", id).RecordNotFound() {
		c.JSON(404, gin.H{
			"status" : "error",
			"message": "record not found",
		})
		c.Abort()
		return 
	}

	config.DB.Model(&item).Where("id = ?", id).Updates(models.Article{
		Title 	: c.PostForm("title"),
		Desc	: c.PostForm("desc"),
		Tag		: c.PostForm("tag"),
	})

	c.JSON(200, gin.H{
		"status" : "berhasil update",
		"message": item,
	})
}