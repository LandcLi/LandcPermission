package permission

import (
	"sync"
	"time"

	"github.com/LandcLi/LandcPermission/model"
)

// defaultEffectiveCacheTTL 是有效权限缓存的默认存活时间。
//
// 选 30s 的理由：这是「权限变更生效延迟」与「数据库/组织服务压力」的折中。
// 写操作会主动失效，所以单实例下变更基本立即生效；跨副本部署时 TTL 就是
// 最大不一致窗口，30s 对多数后台系统可接受。
const defaultEffectiveCacheTTL = 30 * time.Second

// EffectiveCache 缓存用户的「有效权限集合」。
//
// 为什么需要这层缓存：一次 CheckPermission 的成本是
// getAllUserRoleIDs（含一次跨服务 HTTP 调 LandcOrg 取继承角色，默认 3s 超时）
// → role_permissions → permissions，合计 3~5 次 SQL + 1 次 RPC。
// 前端渲染一个菜单要判几十个权限，逐个判定会把权限服务和组织服务一起打满，
// 而组织服务一旦超时，整个权限判定链路都会被拖慢 3s。
//
// 一致性策略：
//   - 写操作（授/撤角色、授/撤直接权限、角色权限变更、权限本身变更）主动失效；
//   - 本缓存是进程内的，多副本部署各副本独立，TTL 即最大不一致窗口；
//   - TTL 设为 0 表示完全关闭缓存（排查「权限改了不生效」时先关它）。
type EffectiveCache struct {
	mu      sync.RWMutex
	ttl     time.Duration
	entries map[uint]effectiveEntry

	// 简单的命中统计，便于观测缓存是否真的起作用。
	hits   uint64
	misses uint64
}

type effectiveEntry struct {
	perms    []model.Permission
	expireAt time.Time
}

// NewEffectiveCache 创建一个 TTL 为 ttl 的缓存；ttl <= 0 表示不缓存。
func NewEffectiveCache(ttl time.Duration) *EffectiveCache {
	return &EffectiveCache{
		ttl:     ttl,
		entries: make(map[uint]effectiveEntry),
	}
}

// Enabled 表示缓存是否生效。
func (c *EffectiveCache) Enabled() bool { return c != nil && c.ttl > 0 }

// Get 取缓存；未命中或已过期返回 ok=false。
func (c *EffectiveCache) Get(userID uint) ([]model.Permission, bool) {
	if !c.Enabled() {
		return nil, false
	}

	c.mu.RLock()
	entry, ok := c.entries[userID]
	c.mu.RUnlock()

	if !ok || time.Now().After(entry.expireAt) {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		if ok {
			// 惰性删除，避免过期条目长期占用内存
			c.mu.Lock()
			delete(c.entries, userID)
			c.mu.Unlock()
		}
		return nil, false
	}

	c.mu.Lock()
	c.hits++
	c.mu.Unlock()
	return entry.perms, true
}

// Set 写入缓存。返回值切片是共享的，调用方不要就地修改。
func (c *EffectiveCache) Set(userID uint, perms []model.Permission) {
	if !c.Enabled() {
		return
	}
	c.mu.Lock()
	c.entries[userID] = effectiveEntry{perms: perms, expireAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

// Invalidate 精确失效某个用户（角色/直接权限变更时调用）。
func (c *EffectiveCache) Invalidate(userID uint) {
	if c == nil {
		return
	}
	c.mu.Lock()
	delete(c.entries, userID)
	c.mu.Unlock()
}

// Clear 全量失效。
// 用于「角色权限」变更：无法廉价反查哪些用户持有该角色，只能整体作废。
func (c *EffectiveCache) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.entries = make(map[uint]effectiveEntry)
	c.mu.Unlock()
}

// Stats 返回命中/未命中次数，供运维观测。
func (c *EffectiveCache) Stats() (hits, misses uint64) {
	if c == nil {
		return 0, 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses
}

// effectiveCache 是进程内单例。
// 用包级变量而不是依赖注入，是因为它属于纯性能优化，不改变任何业务语义，
// 没有必要让上层感知。
var (
	cacheMu       sync.RWMutex
	cacheInstance = NewEffectiveCache(defaultEffectiveCacheTTL)
)

func effectiveCache() *EffectiveCache {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	return cacheInstance
}

// SetEffectiveCacheTTL 设置缓存存活时间，ttl <= 0 关闭缓存。
// 由装配阶段调用（读取配置后），也供测试使用。
func SetEffectiveCacheTTL(ttl time.Duration) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cacheInstance = NewEffectiveCache(ttl)
}

// EffectiveCacheStats 暴露当前缓存命中情况。
func EffectiveCacheStats() (hits, misses uint64, ttl time.Duration) {
	c := effectiveCache()
	hits, misses = c.Stats()
	// ttl 在构造后不再变化，直接读是安全的（避免嵌套加锁）。
	return hits, misses, c.ttl
}
