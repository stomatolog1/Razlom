package servise

// import (
// 	"example/mymodule/internal/model"
// 	"example/mymodule/internal/repository"
// )

// type MusicService struct {
// 	repo *repository.MusicRepository

// }

// func NewMusicrService(repo *repository.MusicRepository) *MusicService{
// 	return &MusicService{repo: repo}
// }

// func (s *MusicService) CreateMusic(req model.NewMusicRequest)(*model.Music, error){
// 	music := &model.Music{
// 		Music_Name: req.Music_Name,
// 		Artist: req.Artist,
// 	}
	
// 	if err:= s.repo.Create(music); err != nil{
// 		return nil, err
// 	}

// 	return music, nil
// }


// func (s *MusicService) GetAllByUser(UserID uint)([]model.Music, error){
// 	return s.repo.GetAllByUser(UserID)
// }

