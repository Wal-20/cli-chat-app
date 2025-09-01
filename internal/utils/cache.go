package utils

import (
	"github.com/patrickmn/go-cache"
	"time"
)

var MembershipCache = cache.New(time.Minute * 5, time.Second * 30)

var AuthCache = cache.New(time.Minute * 5, time.Second) 

var chatroomMessagesCache = cache.New(time.Minute*5, time.Second*30)

var chatroomUsersCache = cache.New(time.Minute*5, time.Second*30)
