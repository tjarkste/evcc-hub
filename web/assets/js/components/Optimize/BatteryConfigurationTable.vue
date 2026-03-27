<template>
	<div class="mb-4">
		<div class="table-responsive">
			<table class="table">
				<thead>
					<tr>
						<th scope="col">{{ $t('hub.debug.battery.battery') }}</th>
						<th scope="col">{{ $t('hub.debug.battery.stateOfCharge') }}</th>
						<th scope="col">{{ $t('hub.debug.battery.socRange') }}</th>
						<th scope="col">{{ $t('hub.debug.battery.energyValue') }}</th>
						<th scope="col">{{ $t('hub.debug.battery.powerRange') }}</th>
						<th scope="col">{{ $t('hub.debug.battery.maxDischarge') }}</th>
						<th scope="col">{{ $t('hub.debug.battery.gridInteraction') }}</th>
						<th scope="col">{{ $t('hub.debug.battery.demandProfile') }}</th>
						<th scope="col">{{ $t('hub.debug.battery.socGoals') }}</th>
					</tr>
				</thead>
				<tbody>
					<tr v-for="(battery, index) in batteries" :key="index">
						<th scope="row">{{ getBatteryTitle(index) }}</th>
						<td>
							<div>{{ formatStateOfCharge(battery.s_initial, index) }}</div>
							<div class="text-muted small">
								{{ formatInitialSocPercentage(battery.s_initial, index) }}
							</div>
						</td>
						<td>
							<div>{{ formatEnergyRange(battery.s_min, battery.s_max) }}</div>
							<div class="text-muted small">
								{{ formatSocRangePercentage(battery.s_min, battery.s_max, index) }}
							</div>
						</td>
						<td>
							<div>{{ formatEnergyValue(battery.p_a) }}</div>
							<div class="text-muted small">
								{{ formatTotalEnergyValue(battery.p_a, index) }}
							</div>
						</td>
						<td>
							{{ formatPowerRange(battery.c_min, battery.c_max) }}
						</td>
						<td>{{ formatPower(battery.d_max) }}</td>
						<td>
							{{ formatGridInteraction(battery) }}
						</td>
						<td>
							<span v-if="battery.p_demand?.length" class="badge bg-info">
								{{ battery.p_demand.length }} {{ $t('hub.debug.battery.steps') }}
							</span>
							<span v-else class="text-muted">{{ $t('hub.debug.battery.none') }}</span>
						</td>
						<td>
							<span v-if="battery.s_goal?.length" class="badge bg-warning">
								{{ battery.s_goal.length }} {{ $t('hub.debug.battery.goals') }}
							</span>
							<span v-else class="text-muted">{{ $t('hub.debug.battery.none') }}</span>
						</td>
					</tr>
				</tbody>
			</table>
		</div>
	</div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import formatter from "@/mixins/formatter";
import type { CURRENCY, BatteryDetail } from "@/types/evcc";

export interface BatteryConfig {
	c_min: number;
	c_max: number;
	d_max: number;
	s_min: number;
	s_max: number;
	s_initial: number;
	p_a: number;
	charge_from_grid?: boolean;
	discharge_to_grid?: boolean;
	p_demand?: number[];
	s_goal?: number[];
}

export default defineComponent({
	name: "BatteryConfigurationTable",
	mixins: [formatter],
	props: {
		batteries: {
			type: Array as PropType<BatteryConfig[]>,
			required: true,
		},
		batteryDetails: {
			type: Array as PropType<BatteryDetail[]>,
			required: true,
		},
		currency: {
			type: String as PropType<CURRENCY>,
			required: true,
		},
	},
	methods: {
		formatPower(watts: number): string {
			return this.fmtW(watts, this.POWER_UNIT.KW, true, 1);
		},
		formatEnergy(wh: number): string {
			return this.fmtWh(wh, this.POWER_UNIT.KW, true, 1);
		},
		formatPowerRange(min: number, max: number): string {
			const minValue = this.fmtW(min, this.POWER_UNIT.KW, false, 1);
			const maxValue = this.fmtW(max, this.POWER_UNIT.KW, true, 1);
			return `${minValue} – ${maxValue}`;
		},
		formatEnergyRange(min: number, max: number): string {
			const minValue = this.fmtWh(min, this.POWER_UNIT.KW, false, 1);
			const maxValue = this.fmtWh(max, this.POWER_UNIT.KW, true, 1);
			return `${minValue} – ${maxValue}`;
		},
		formatEnergyValue(valuePerWh: number): string {
			return this.fmtPricePerKWh(valuePerWh * 1000, this.currency, false, true);
		},
		getBatteryTitle(index: number): string {
			const detail = this.batteryDetails[index];
			return detail ? detail.title || detail.name : `Battery ${index + 1}`;
		},
		formatStateOfCharge(initialSocWh: number, index: number): string {
			const detail = this.batteryDetails[index];
			if (detail?.capacity) {
				const initialSocKWh = this.fmtWh(initialSocWh, this.POWER_UNIT.KW, false, 1);
				const capacityKWh = this.fmtWh(detail.capacity * 1000, this.POWER_UNIT.KW, true, 1);
				return `${initialSocKWh} ${this.$t('hub.debug.of')} ${capacityKWh}`;
			}
			return "-";
		},
		formatInitialSocPercentage(initialSocWh: number, index: number): string {
			const detail = this.batteryDetails[index];
			if (detail?.capacity && detail.capacity > 0) {
				const percentage = (initialSocWh / 1000 / detail.capacity) * 100;
				return this.fmtPercentage(percentage, 0);
			}
			return "";
		},
		formatSocRangePercentage(minSocWh: number, maxSocWh: number, index: number): string {
			const detail = this.batteryDetails[index];
			if (detail?.capacity && detail.capacity > 0) {
				const minPercentage = (minSocWh / 1000 / detail.capacity) * 100;
				const maxPercentage = (maxSocWh / 1000 / detail.capacity) * 100;
				return `${this.fmtPercentage(minPercentage, 0)} – ${this.fmtPercentage(maxPercentage, 0)}`;
			}
			return "";
		},
		formatTotalEnergyValue(valuePerWh: number, index: number): string {
			const detail = this.batteryDetails[index];
			if (detail?.capacity && detail.capacity > 0) {
				const totalValue = valuePerWh * detail.capacity * 1000; // Convert kWh to Wh for calculation
				return this.fmtMoney(totalValue, this.currency, true, true);
			}
			return "";
		},
		formatGridInteraction(battery: BatteryConfig): string {
			const canCharge = battery.charge_from_grid;
			const canDischarge = battery.discharge_to_grid;

			if (canCharge && canDischarge) {
				return this.$t("hub.debug.battery.chargeDischarge");
			} else if (canCharge) {
				return this.$t("hub.debug.battery.charge");
			} else if (canDischarge) {
				return this.$t("hub.debug.battery.discharge");
			} else {
				return this.$t("hub.debug.battery.none");
			}
		},
	},
});
</script>

<style scoped>
.table td,
.table th {
	font-variant-numeric: tabular-nums;
}
</style>
