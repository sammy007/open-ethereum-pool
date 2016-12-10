import Ember from 'ember';

export default Ember.Component.extend({

  summaryOptions: {
    chart: {
        plotBackgroundColor: null,
        plotBorderWidth: null,
        plotShadow: false,
        type: 'pie'
    },
    title: {
        text: 'Total weight of gear in each category'
    },
    tooltip: {
        pointFormat: '<b>{point.percentage:.1f}%</b> of {series.name}'
    },
    plotOptions: {
        pie: {
            allowPointSelect: true,
            cursor: 'pointer',
            dataLabels: {
                enabled: false
            },
            showInLegend: true
        }
    }
  },
  summaryData: [{
      name: 'gear',
      colorByPoint: true,
      data: [
          {y: 10, name: 'Test1'},
          {y: 12, name: 'Test2'},
          {y: 40, name: 'Test3'}
          ]
  }]

});
