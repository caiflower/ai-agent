package bean

import "time"

type BaseModel struct {
	Id         int       //主键
	CreateTime time.Time //创建时间
	UpdateTime time.Time //更新时间
	Status     int       //状态
}
