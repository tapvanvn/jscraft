package entity

import (
	"fmt"
	"strconv"
	"sync"
)

var __require_id int = 0

//RequireTable save require files to build a file
type RequireTable struct {
	OwnerFile  string
	Files      map[string]bool
	file_mux   sync.Mutex
	lock_count int
	id         int
}

func CreateRequireTable(filePath string) *RequireTable {
	table := RequireTable{OwnerFile: filePath}
	table.Init()
	return &table
}

//Init init
func (table *RequireTable) Init() {
	table.id = __require_id
	__require_id++
	table.Files = make(map[string]bool)
}

//AddFile add file
func (table *RequireTable) AddFile(filePath string) {
	table.lock_count++
	//fmt.Printf("lock %d add file: %d %s\n", table.id, table.lock_count, table.OwnerFile)
	table.file_mux.Lock()

	table.Files[filePath] = true
	fmt.Println("\n" + table.OwnerFile + "\n===>" + filePath + strconv.Itoa(len(table.Files)) + "\n")

	table.lock_count--
	//fmt.Printf("unlock %d add file: %d \n", table.id, table.lock_count)
	table.file_mux.Unlock()
}
