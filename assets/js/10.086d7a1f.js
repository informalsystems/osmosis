(window.webpackJsonp=window.webpackJsonp||[]).push([[10],{439:function(e,t,i){e.exports=i.p+"assets/img/lbp.a7971717.png"},440:function(e,t,i){e.exports=i.p+"assets/img/gauges.7b963447.png"},492:function(e,t,i){"use strict";i.r(t);var a=i(8),o=Object(a.a)({},(function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("ContentSlotsDistributor",{attrs:{"slot-key":e.$parent.slotKey}},[a("h1",{attrs:{id:"learn-more"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#learn-more"}},[e._v("#")]),e._v(" Learn More")]),e._v(" "),a("h2",{attrs:{id:"liquidity-bootstrapping-pools"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#liquidity-bootstrapping-pools"}},[e._v("#")]),e._v(" Liquidity Bootstrapping Pools")]),e._v(" "),a("p",[e._v("Osmosis offers a convenient design for Liquidity Bootstrapping Pools (LBPs), a special type of AMM designed for token sales. New protocols can use Osmosis’ LBP feature to distribute tokens and achieve initial price discovery.")]),e._v(" "),a("p",[e._v("LBPs differ from other liquidity pools in terms of the ratio of assets within the pool. In LBPs, the ratio adjusts over time. LBPs involve an initial ratio, a target ratio, and an interval of time over which the ratio adjusts. These weights are customizable before launch. One can create an LBP with an initial ratio of 90-10, with the goal of reaching 50-50 over a month. The ratio continues to gradually adjust until the weights are equal within the pool. Like traditional LPs, the prices of assets within the pool is based on the ratio at the time of trade.")]),e._v(" "),a("p",[e._v("After the LBP period ends or the final ratio is reached, the pool converts into a traditional LP pool.")]),e._v(" "),a("p",[e._v("LBPs facilitate price discovery by demonstrating the acceptable market price of an asset. Ideally, LBPs will have very few buyers at the time of launch. The price slowly declines until traders are willing to step in and buy the asset. Arbitrage maintains this price for the remainder of the LBP. The process is shown by the blue line below. The price declines as the ratios adjust. Buyers step in until the acceptable price is again reached.")]),e._v(" "),a("p",[a("img",{attrs:{src:i(439),alt:""}})]),e._v(" "),a("p",[e._v("Choosing the correct parameters is very important to price discovery for an LBP. If the initial price is too low, the asset will get bought up as soon as the pool is launched. It is also possible that the targeted final price is too high, creating little demand for the asset. The green line above shows this scenario. A buyer places a large order, but afterwards the price continues to decline as no additional buyers are willing to bite.")]),e._v(" "),a("p",[e._v("Osmosis aims to provide an intuitive, easy-to-use LBP design to give protocols the best chance of a successful sale. The LBP feature facilitates initial price discovery for tokens and allows protocols to fairly distribute tokens to project stakeholders.")]),e._v(" "),a("h2",{attrs:{id:"bonded-liquidity-gauges"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#bonded-liquidity-gauges"}},[e._v("#")]),e._v(" Bonded Liquidity Gauges")]),e._v(" "),a("p",[e._v("Bonded Liquidity Gauges are mechanisms for distributing liquidity incentives to LP tokens that have been bonded for a minimum amount of time. 45% of the daily issuance of OSMO goes towards these liquidity incentives.")]),e._v(" "),a("p",[e._v("For instance, a Pool 1 LP share, 1-week gauge would distribute rewards to users who have bonded Pool1 LP tokens for one week or longer. The amount that each user receives is in proportion to the number of their bonded tokens.")]),e._v(" "),a("p",[e._v("A bonded LP position can be eligible for multiple gauges. Qualifications for a gauge only involve a minimum bonding time.")]),e._v(" "),a("p",[a("img",{attrs:{src:i(440),alt:"Tux, the Linux mascot"}})]),e._v(" "),a("p",[e._v("The rewards earned from liquidity mining are not subject to unbonding. Rewards are liquid and transferable immediately. Only the principal bonded shares are subject to the unbonding period.")]),e._v(" "),a("h2",{attrs:{id:"allocation-points"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#allocation-points"}},[e._v("#")]),e._v(" Allocation Points")]),e._v(" "),a("p",[e._v("Not all pools have incentivized gauges. In Osmosis, staked OSMO holders choose which pools to incentivize via on-chain governance proposals. To incentivize a pool, governance can assign “allocation points” to specific gauges. At the end of every daily epoch, 45% of the newly released OSMO (the portions designated for liquidity incentives) is distributed proportionally to the allocation points that each gauge has. The percent of the OSMO liquidity rewards that each gauge receives is calculated as its number of points divided by the total number of allocation points.")]),e._v(" "),a("p",[e._v("Take, for example, a scenario in which three gauges are incentivized:")]),e._v(" "),a("ul",[a("li",[e._v("Gauge #3 – 10 allocation points")]),e._v(" "),a("li",[e._v("Gauge #4 – 5 allocation points")]),e._v(" "),a("li",[e._v("Gauge #7 – 5 allocation points")])]),e._v(" "),a("p",[e._v("20 total allocation points are assigned in this scenario. At the end of the daily epochs, Gauge #3 will receive 50% (10 out of 20) of the liquidity incentives minted. Gauges #4 and #7 will receive 25% each.")]),e._v(" "),a("p",[e._v("Governance can pass an UpdatePoolIncentives proposal to edit the existing allocation points of any gauge. By setting a gauge’s allocation to zero, it can remove it from the list of incentivized gauges entirely. Proposals can also set the allocation points of new gauges. When a new gauge is added, the total number of allocation points increases, thus diluting all the existing incentivized gauges.\nGauge #0 is a special gauge that sends its incentives directly to the chain community pool. Assigning allocation points to gauge #0 allows governance to save some of the current liquidity mining incentives to be spent at a later time.")]),e._v(" "),a("p",[e._v("At genesis, the only gauge that will be incentivized is Gauge #0, (the community pool gauge). However, a governance proposal can come immediately after launch to choose which gauges/pools to incentivize. Governance voting period at launch is only 3 days at launch, so liquidity incentives may be activated as soon as 3 days after genesis.")]),e._v(" "),a("h2",{attrs:{id:"external-incentives"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#external-incentives"}},[e._v("#")]),e._v(" External Incentives")]),e._v(" "),a("p",[e._v("Osmosis not only allows the community to add incentives to gauges. Anyone can deposit tokens into a gauge to be distributed. This feature allows outside parties to augment Osmosis’ own liquidity incentive program.")]),e._v(" "),a("p",[e._v("For example, there may be an ATOM<>FOOCOIN pool that has a one-day gauge incentivized by governance OSMO rewards. However, the Foo Foundation may also choose to add additional incentives to the one-day gauge or even add incentives to a new gauge (such as one-week gauge).")]),e._v(" "),a("p",[e._v("These external incentive providers can also set up long-lasting incentive programs that distribute rewards over an extended time period. For example, the Foo Foundation can deposit 30,000 Foocoins to be distributed over a one-month liquidity program. The program will automatically distribute 1000 Foocoins per day to the gauge.")]),e._v(" "),a("h2",{attrs:{id:"fees"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#fees"}},[e._v("#")]),e._v(" Fees")]),e._v(" "),a("p",[e._v("In addition to liquidity mining, Osmosis provides three sources of revenue: transaction fees, swap fees, and exit fees.")]),e._v(" "),a("h3",{attrs:{id:"tx-fees"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#tx-fees"}},[e._v("#")]),e._v(" TX Fees")]),e._v(" "),a("p",[e._v("Transaction fees are paid by any user to post a transaction on the chain. The fee amount is determined by the computation and storage costs of the transaction. Minimum gas costs are determined by the proposer of a block in which the transaction is included. This transaction fee is distributed to OSMO stakers on the network.\nValidators can choose which assets to accept for fees in the blocks that they propose. This optionality is a unique feature of Osmosis.")]),e._v(" "),a("h3",{attrs:{id:"swap-fees"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#swap-fees"}},[e._v("#")]),e._v(" Swap Fees")]),e._v(" "),a("p",[e._v("Swap fees are fees charged for making a swap in an LP pool. The fee is paid by the trader in the form of the input asset. Pool creators specify the swap fee when establishing the pool. The total fee for a particular trade is calculated as percentage of swap size. Fees are added to the pool, effectively resulting in pro-rata distribution to LPs proportional to their share of the total pool.")]),e._v(" "),a("h3",{attrs:{id:"exit-fees"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#exit-fees"}},[e._v("#")]),e._v(" Exit Fees")]),e._v(" "),a("p",[e._v("Osmosis LPs pay a small fee when withdrawing from the pool. Similar to swap fees, exit fees per pool are set by the pool creator.\nExit fees are paid in LP tokens. Users withdraw their tokens, minus a percent for the exit fee. These LP shares are burned, resulting in pro-rata distribution to remaining LPs.")])])}),[],!1,null,null,null);t.default=o.exports}}]);