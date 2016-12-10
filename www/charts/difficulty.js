import Highcharts from 'ember-highcharts/components/high-charts';

export default Highcharts.extend({
   chartMode: '', // empty, 'StockChart', or 'Map'
   chartOptions: {
     chart: {
       type: 'bar'
     },
     title: {
       text: 'Fruit Consumption'
     },
     xAxis: {
       categories: ['Apples', 'Bananas', 'Oranges']
     },
     yAxis: {
       title: {
         text: 'Fruit eaten'
       }
     }
   },
   chartData: [{
     name: 'Jane',
     data: [1, 0, 4]
   }, {
     name: 'John',
     data: [5, 7, 3]
   }],
   theme: "defaultTheme"
});
