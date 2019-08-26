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
	"strings"
)

// trolley 购物车
//
// 这里面能存放很多手提袋
//
// purses 手提袋集合
//
// b+tree 模型 degree=128;level=4;nodes=[d^l]/(d-1)=2113665;
//
// node 内范围控制数量 key=127
//
// tree 内范围控制数量 n*k=268435455
type trolley struct {
	key         uint32
	realKey     uint32
	flexibleKey uint32
	mall        *mall
	purses      []*purse
}

func (t *trolley) put(originalKey Key, key uint32, value interface{}) error {
	t.realKey = t.mall.flexibleKey / purseDistance
	t.flexibleKey = t.mall.flexibleKey - t.realKey*purseDistance
	//log.Self.Debug("trolley", log.Uint32("key", key), log.Uint32("realKey", realKey))
	t.createChild(t.realKey)
	return t.purses[t.realKey].put(originalKey, key, value)
}

func (t *trolley) get(originalKey Key, key uint32) (interface{}, error) {
	t.realKey = t.mall.flexibleKey / purseDistance
	t.flexibleKey = t.mall.flexibleKey - t.realKey*purseDistance
	if t.existChild(t.realKey) {
		return t.purses[t.realKey].get(originalKey, key)
	} else {
		return nil, errors.New(strings.Join([]string{"trolley key", string(originalKey), "is nil"}, " "))
	}
}

func (t *trolley) existChild(index uint32) bool {
	return nil != t.purses[index]
}

func (t *trolley) createChild(index uint32) {
	if !t.existChild(index) {
		t.purses[index] = &purse{
			key:     t.mall.city.realKey*mallDistance + t.mall.realKey*trolleyDistance + index*purseDistance + purseDistance,
			trolley: t,
			box:     make([]*box, boxCount),
		}
	}
}
