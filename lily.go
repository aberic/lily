/*
 * Copyright (c) 2019. Aberic - All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package Lily

import (
	"errors"
	"github.com/ennoo/rivet/utils/cryptos"
	str "github.com/ennoo/rivet/utils/string"
	"strings"
	"sync"
)

const (
	sysCheckbook           = "lily"
	userShopper            = "_user"
	userSequenceShopper    = "_user_id"
	cardShopper            = "_card"
	cardSequenceShopper    = "_card_id"
	defaultShopper         = "_default"
	defaultSequenceShopper = "_default_id"
)

var (
	lilyInstance      *Lily
	onceLily          sync.Once
	checkbookExistErr = errors.New("checkbook(database) already exist")          // checkbookExistErr 自定义error信息
	shopperExistErr   = errors.New("shopper(form) already exist")                // shopperExistErr 自定义error信息
	errorDataIsNil    = errors.New("checkbook(database) had never been created") // errorDataIsNil 自定义error信息
)

// Lily 祖宗！
//
// 全库唯一常住内存对象，并持有所有库的句柄
//
// API 入口
//
// 存储格式 {dataDir}/checkbook/{dataName}/{shopperName}/{shopperName}.dat/idx...
type Lily struct {
	defaultCheckbook *checkbook
	checkbooks       map[string]*checkbook
	once             sync.Once
}

// ObtainLily 获取 Lily 对象
//
// 会初始化一个空 Lily，如果是第一次调用的话
//
// 首次调用后需要执行 initialize() 初始化方法
//
// 或者通过外部调用 Start() 来执行初始化操作
//
// 调用 Restart() 会恢复 Lily 的索引，如果 Lily 索引存在，则 Restart() 什么也不会做
//
// 会返回一个已创建的 Lily，如果非第一次调用的话
func ObtainLily() *Lily {
	onceLily.Do(func() {
		lilyInstance = &Lily{
			checkbooks: map[string]*checkbook{},
		}
	})
	return lilyInstance
}

// Start 启动lily
//
// 调用后执行 initialize() 初始化方法
func (l *Lily) Start() {
	l.initialize()
}

// Restart 重新启动lily
//
// 调用 Restart() 会恢复 Lily 的索引，如果 Lily 索引存在，则 Restart() 什么也不会做
func (l *Lily) Restart() {

}

// initialize 初始化默认库及默认表
func (l *Lily) initialize() {
	l.once.Do(func() {
		data, err := l.CreateCheckbook(sysCheckbook)
		if nil != err {
			if err == fileExistErr {
				l.Restart()
				return
			} else {
				panic(err)
			}
		}
		if err = data.createShopper(userShopper, "default checkbook shopper", true); nil != err {
			_ = rmDataPath(sysCheckbook)
			return
		}
		if err = data.createShopper(cardShopper, "default checkbook shopper", true); nil != err {
			_ = rmDataPath(sysCheckbook)
			return
		}
		if err = data.createShopper(defaultShopper, "default checkbook shopper", true); nil != err {
			_ = rmDataPath(sysCheckbook)
			return
		}
		l.defaultCheckbook = data
	})
}

// CreateCheckbook 新建数据库
//
// 新建数据库会同时创建一个名为_default的表，未指定表明的情况下使用put/get等方法会操作该表
//
// name 数据库名称
func (l *Lily) CreateCheckbook(name string) (*checkbook, error) {
	// 确定库名不重复
	for k := range l.checkbooks {
		if k == name {
			return nil, checkbookExistErr
		}
	}
	// 确保数据库唯一ID不重复
	id := l.name2id(name)
	if err := mkDataPath(id); nil != err {
		return nil, err
	}
	data := &checkbook{name: name, id: id, shoppers: map[string]*shopper{}}
	l.checkbooks[name] = data
	return data, nil
}

// CreateShopper 创建表
//
// name 表名称
//
// comment 表描述
//
// sequence 是否启用自增ID索引
func (l *Lily) CreateShopper(checkbookName, shopperName, comment string, sequence bool) error {
	if cb := l.checkbooks[checkbookName]; nil != cb {
		return cb.createShopper(shopperName, comment, sequence)
	}
	return errorDataIsNil
}

// Put 新增数据
//
// 向_default表中新增一条数据，key相同则覆盖
//
// key 插入数据唯一key
//
// value 插入数据对象
//
// 返回 hashKey
func (l *Lily) Put(key Key, value interface{}) (uint32, error) {
	return l.defaultCheckbook.Insert(defaultShopper, key, value)
}

// Get 获取数据
//
// 向_default表中查询一条数据并返回
//
// key 插入数据唯一key
func (l *Lily) Get(key Key) (interface{}, error) {
	return l.defaultCheckbook.Query(defaultShopper, key)
}

// name2id 确保数据库唯一ID不重复
func (l *Lily) name2id(name string) string {
	id := cryptos.MD516(name)
	have := true
	for have {
		have = false
		for _, v := range l.checkbooks {
			if v.id == id {
				have = true
				id = cryptos.MD516(strings.Join([]string{id, str.RandSeq(3)}, ""))
				break
			}
		}
	}
	return id
}
