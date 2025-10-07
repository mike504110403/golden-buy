<template>
  <div class="dashboard">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">儀表板</h1>

    <!-- 價格卡片網格 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <price-card
        v-for="metal in ALL_METALS"
        :key="metal"
        :symbol="metal"
      />
    </div>

    <!-- K 線圖表 -->
    <div class="card">
      <h2 class="text-xl font-bold text-gray-900 mb-4">價格走勢</h2>
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <KlineChart symbol="GOLD" />
        <KlineChart symbol="SILVER" />
        <KlineChart symbol="PLATINUM" />
        <KlineChart symbol="PALLADIUM" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ALL_METALS } from '../../utils/constants'
import PriceCard from '../../components/common/PriceCard.vue'
import KlineChart from '../../components/charts/KlineChart.vue'
</script>


