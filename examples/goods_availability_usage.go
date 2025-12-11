package examples

import (
	"context"
	"fmt"
	dao "teachat/DAO"
)

// 使用示例：如何管理物资在不同组织中的可用性状态

func ExampleGoodsAvailabilityUsage() {
	ctx := context.Background()

	// 1. 创建一个物资
	goods := &dao.Goods{
		RecorderUserId:   1,
		Name:             "笔记本电脑",
		Nickname:         "工作本",
		Category:         dao.GoodsCategoryPhysical,
		PhysicalState:    dao.PhysicalNew,
		OperationalState: dao.OperationalNormal,
		// 注意：不再有 Availability 字段
	}

	err := goods.Create(ctx)
	if err != nil {
		fmt.Printf("创建物资失败: %v\n", err)
		return
	}

	// 2. 将物资分配给家庭，状态为"可用"
	goodsFamily := &dao.GoodsFamily{
		FamilyId:     1,
		GoodsId:      goods.Id,
		Availability: dao.Available,
	}

	err = goodsFamily.Create(ctx)
	if err != nil {
		fmt.Printf("创建家庭物资关系失败: %v\n", err)
		return
	}

	// 3. 将同一物资分配给团队，状态为"使用中"
	goodsTeam := &dao.GoodsTeam{
		TeamId:       1,
		GoodsId:      goods.Id,
		Availability: dao.InUse,
	}

	err = goodsTeam.Create(ctx)
	if err != nil {
		fmt.Printf("创建团队物资关系失败: %v\n", err)
		return
	}

	// 4. 查询物资在家庭中的状态
	familyRelation, err := dao.GetGoodsFamilyByIds(1, goods.Id, ctx)
	if err != nil {
		fmt.Printf("查询家庭物资关系失败: %v\n", err)
		return
	}

	if familyRelation != nil {
		fmt.Printf("物资在家庭中的状态: %s\n", dao.GoodsAvailabilityString(familyRelation.Availability))
	}

	// 5. 查询物资在团队中的状态
	teamRelation, err := dao.GetGoodsTeamByIds(1, goods.Id, ctx)
	if err != nil {
		fmt.Printf("查询团队物资关系失败: %v\n", err)
		return
	}

	if teamRelation != nil {
		fmt.Printf("物资在团队中的状态: %s\n", dao.GoodsAvailabilityString(teamRelation.Availability))
	}

	// 6. 更新物资在家庭中的状态
	familyRelation.Availability = dao.InUse
	err = familyRelation.UpdateAvailability(ctx)
	if err != nil {
		fmt.Printf("更新家庭物资状态失败: %v\n", err)
		return
	}

	// 7. 获取家庭的所有物资
	familyGoods, availabilities, err := dao.GetGoodsByFamilyId(1, ctx)
	if err != nil {
		fmt.Printf("获取家庭物资失败: %v\n", err)
		return
	}

	fmt.Printf("家庭拥有 %d 个物资:\n", len(familyGoods))
	for i, g := range familyGoods {
		fmt.Printf("- %s: %s\n", g.Name, dao.GoodsAvailabilityString(availabilities[i]))
	}
}
