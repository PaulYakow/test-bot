package controller

import (
	"github.com/vitaliy-ukiru/fsm-telebot"
	tele "gopkg.in/telebot.v3"
)

type User struct {
	service UserService
	view    UserView
}

func NewUser(service UserService, view UserView) *User {
	return &User{
		service: service,
		view:    view,
	}
}

type Absence struct {
	service AbsenceService
	view    AbsenceView
}

func NewAbsence(service AbsenceService, view AbsenceView) *Absence {
	return &Absence{
		service: service,
		view:    view,
	}
}

type Set struct {
	User    *User
	Absence *Absence
}

type Controller struct {
	bot     *tele.Bot
	manager *fsm.Manager
	user    *User
	absence *Absence
}

// TODO: вариант формирования набора. Можно вынести создание структур Service и View в соответствующие пакеты.

type Service struct {
	user    UserService
	absence AbsenceService
}

type View struct {
	user    UserView
	absence AbsenceView
}

type set struct {
	service *Service
	view    *View
}

func NewService(us UserService, as AbsenceService) *Service {
	return &Service{
		user:    us,
		absence: as,
	}
}

func NewView(uv UserView, av AbsenceView) *View {
	return &View{
		user:    uv,
		absence: av,
	}
}
