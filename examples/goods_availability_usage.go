package examples

import (
	"context"
	"fmt"
	data "teachat/DAO"
)

// 使用示例：如何管理物资在不同组织中的可用性状态

func ExampleGoodsAvailabilityUsage() {
	ctx := context.Background()

	// 1. 创建一个物资
	goods := &data.Goods{
		RecorderUserId: 1,
		Name:          "笔记本电脑",
		Nickname:      "工作本",
		Category:      data.GoodsCategoryPhysical,
		PhysicalState: data.PhysicalNew,
		OperationalState: data.OperationalNormal,
		// 注意：不再有 Availability 字段
	}
	
	err := goods.Create(ctx)
	if err != nil {
		fmt.Printf("创建物资失败: %v\n", err)
		return
	}

	// 2. 将物资分配给家庭，状态为"可用"
	goodsFamily := &data.GoodsFamily{
		FamilyId:     1,
		GoodsId:      goods.Id,
		Availability: data.Available,
	}
	
	err = goodsFamily.Create(ctx)
	if err != nil {
		fmt.Printf("创建家庭物资关系失败: %v\n", err)
		return
	}

	// 3. 将同一物资分配给团队，状态为"使用中"
	goodsTeam := &data.GoodsTeam{
		TeamId:       1,
		GoodsId:      goods.Id,
		Availability: data.InUse,
	}
	
	err = goodsTeam.Create(ctx)
	if err != nil {
		fmt.Printf("创建团队物资关系失败: %v\n", err)
		return
	}

	// 4. 查询物资在家庭中的状态
	familyRelation, err := data.GetGoodsFamilyByIds(1, goods.Id, ctx)
	if err != nil {
		fmt.Printf("查询家庭物资关系失败: %v\n", err)
		return
	}
	
	if familyRelation != nil {
		fmt.Printf("物资在家庭中的状态: %s\n", data.AvailabilityString(familyRelation.Availability))
	}

	// 5. 查询物资在团队中的状态
	teamRelation, err := data.GetGoodsTeamByIds(1, goods.Id, ctx)
	if err != nil {
		fmt.Printf("查询团队物资关系失败: %v\n", err)
		return
	}
	
	if teamRelation != nil {
		fmt.Printf("物资在团队中的状态: %s\n", data.AvailabilityString(teamRelation.Availability))
	}

	// 6. 更新物资在家庭中的状态
	familyRelation.Availability = data.InUse
	err = familyRelation.UpdateAvailability(ctx)
	if err != nil {
		fmt.Printf("更新家庭物资状态失败: %v\n", err)
		return
	}

	// 7. 获取家庭的所有物资
	familyGoods, availabilities, err := data.GetGoodsByFamilyId(1, ctx)
	if err != nil {
		fmt.Printf("获取家庭物资失败: %v\n", err)
		return
	}
	
	fmt.Printf("家庭拥有 %d 个物资:\n", len(familyGoods))
	for i, g := range familyGoods {
		fmt.Printf("- %s: %s\n", g.Name, data.AvailabilityString(availabilities[i]))
	}
}