// Package novel re-exports domain types for use in the service layer.
package novel

import (
	"github.com/835-droid/ms-ai-backend/internal/domain/novel"
)

// Re-export domain types for service layer use
type (
	Novel                  = novel.Novel
	RankedNovel            = novel.RankedNovel
	NovelChapter           = novel.NovelChapter
	NovelRating            = novel.NovelRating
	ChapterRating          = novel.ChapterRating
	ReactionType           = novel.ReactionType
	NovelReaction          = novel.NovelReaction
	UserFavorite           = novel.UserFavorite
	FavoriteList           = novel.FavoriteList
	FavoriteListItem       = novel.FavoriteListItem
	NovelComment           = novel.NovelComment
	ChapterComment         = novel.ChapterComment
	ChapterCommentReaction = novel.ChapterCommentReaction
	ReadingProgress        = novel.ReadingProgress
	ViewingHistory         = novel.ViewingHistory
)

// Reaction type constants
const (
	ReactionUpvote    = novel.ReactionUpvote
	ReactionFunny     = novel.ReactionFunny
	ReactionLove      = novel.ReactionLove
	ReactionSurprised = novel.ReactionSurprised
	ReactionAngry     = novel.ReactionAngry
	ReactionSad       = novel.ReactionSad
)

// Repository interfaces
type (
	NovelRepository                = novel.NovelRepository
	NovelChapterRepository         = novel.NovelChapterRepository
	NovelFavoriteListRepository    = novel.NovelFavoriteListRepository
	NovelReadingProgressRepository = novel.NovelReadingProgressRepository
	NovelViewingHistoryRepository  = novel.NovelViewingHistoryRepository
)
