package ds

type User struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Username    string `gorm:"type:varchar(50);unique;not null" json:"username"`
	Password    string `gorm:"type:varchar(255);not null" json:"-"`
	IsModerator bool   `gorm:"type:boolean;default:false" json:"is_moderator"`
}
