package main

import (
	git "github.com/yi-nology/git-manage-service/biz/kitex_gen/git/gitservice"
	"log"
)

func main() {
	svr := git.NewServer(new(GitServiceImpl))

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
