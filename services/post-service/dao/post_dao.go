package dao

import (
	"post-service/model"

	"gorm.io/gorm"
)

// CreatePost 创建动态
func CreatePost(post *model.Post) error {
	return DB.Create(post).Error
}

// GetPostByID 根据动态ID获取动态信息
func GetPostByID(postID uint) (*model.Post, error) {
	var post model.Post
	err := DB.Where("id = ?", postID).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetPostsByUserID 根据用户ID获取动态列表
func GetPostsByUserID(userID uint, page, pageSize int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64

	// 计算总数
	if err := DB.Model(&model.Post{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&posts).Error

	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetAllPosts 获取所有动态（分页）
func GetAllPosts(page, pageSize int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64

	// 计算总数
	if err := DB.Model(&model.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := DB.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&posts).Error

	if err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// UpdatePost 更新动态
func UpdatePost(post *model.Post) error {
	return DB.Save(post).Error
}

// DeletePost 删除动态（软删除）
func DeletePost(postID uint) error {
	return DB.Delete(&model.Post{}, postID).Error
}

// IncrementLikeCount 增加点赞数
func IncrementLikeCount(postID uint) error {
	return DB.Model(&model.Post{}).Where("id = ?", postID).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
}

// DecrementLikeCount 减少点赞数
func DecrementLikeCount(postID uint) error {
	return DB.Model(&model.Post{}).Where("id = ?", postID).
		UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error
}

// IncrementFavoriteCount 增加收藏数
func IncrementFavoriteCount(postID uint) error {
	return DB.Model(&model.Post{}).Where("id = ?", postID).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1)).Error
}

// DecrementFavoriteCount 减少收藏数
func DecrementFavoriteCount(postID uint) error {
	return DB.Model(&model.Post{}).Where("id = ?", postID).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1)).Error
}

// IncrementShareCount 增加分享数
func IncrementShareCount(postID uint) error {
	return DB.Model(&model.Post{}).Where("id = ?", postID).
		UpdateColumn("share_count", gorm.Expr("share_count + ?", 1)).Error
}

// DecrementShareCount 减少分享数
func DecrementShareCount(postID uint) error {
	return DB.Model(&model.Post{}).Where("id = ?", postID).
		UpdateColumn("share_count", gorm.Expr("share_count - ?", 1)).Error
}

// IncrementDislikeCount 增加踩数
func IncrementDislikeCount(postID uint) error {
	return DB.Model(&model.Post{}).Where("id = ?", postID).
		UpdateColumn("dislike_count", gorm.Expr("dislike_count + ?", 1)).Error
}

// DecrementDislikeCount 减少踩数
func DecrementDislikeCount(postID uint) error {
	return DB.Model(&model.Post{}).Where("id = ?", postID).
		UpdateColumn("dislike_count", gorm.Expr("dislike_count - ?", 1)).Error
}

// CheckUserLikedPost 检查用户是否点赞了某个动态
func CheckUserLikedPost(userID, postID uint) (bool, error) {
	var count int64
	err := DB.Model(&model.Like{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error
	return count > 0, err
}

// CheckUserDislikedPost 检查用户是否踩了某个动态
func CheckUserDislikedPost(userID, postID uint) (bool, error) {
	var count int64
	err := DB.Model(&model.Dislike{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error
	return count > 0, err
}

// CheckUserFavoritedPost 检查用户是否收藏了某个动态
func CheckUserFavoritedPost(userID, postID uint) (bool, error) {
	var count int64
	err := DB.Model(&model.Favorite{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error
	return count > 0, err
}

// CreateLike 创建点赞记录
func CreateLike(like *model.Like) error {
	return DB.Create(like).Error
}

// CreateFavorite 创建收藏记录
func CreateFavorite(favorite *model.Favorite) error {
	return DB.Create(favorite).Error
}

// DeleteLike 删除点赞记录
func DeleteLike(userID, postID uint) error {
	return DB.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.Like{}).Error
}

// DeleteFavorite 删除收藏记录
func DeleteFavorite(userID, postID uint) error {
	return DB.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.Favorite{}).Error
}

// CreateDislike 创建踩记录
func CreateDislike(dislike *model.Dislike) error {
	return DB.Create(dislike).Error
}

// DeleteDislike 删除踩记录
func DeleteDislike(userID, postID uint) error {
	return DB.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.Dislike{}).Error
}
