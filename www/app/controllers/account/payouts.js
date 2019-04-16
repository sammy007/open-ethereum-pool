import Controller from '@ember/controller';
import { inject } from '@ember/controller';
import { inject as service} from '@ember/service';
import { computed } from '@ember/object';

export default Controller.extend({
  applicationController: inject('application'),
  config: computed.reads('applicationController.config'),
  stats: computed.reads('applicationController.model.stats'),
  intl: service(),

  chartPaymentText: computed('model', {
    get() {
      var outText = this.get('model.paymentCharts');
      if (!outText) {
        return 0;
      }
      return outText;
    }
  }),

  chartPayment: computed('intl', 'model.paymentCharts', {
    get() {
        var e = this,
            t = e.getWithDefault("model.paymentCharts", []),
            a = {
                chart: {
                    backgroundColor: "rgba(255, 255, 255, 0.1)",
                    type: "column",
                    marginRight: 10,
                    height: 200,
                    events: {
                        load: function() {
                            var self = this;
                            var chartInterval = setInterval(function() {
                                if (!self.series) {
                                    clearInterval(chartInterval);
                                    return;
                                }
                                t = e.getWithDefault("model.paymentCharts", []);
                                var data = [];
                                t.forEach(function(d) {
                                    var r = new Date(1000 * d.x);
                                    var l = r.toLocaleString();
                                    var n = d.amount / 1000000000;
                                    data.push({x: r, d: l, y: n});
                                });
                                if (data.length > 0) {
                                    self.series[0].setData(data, true, {}, true);
                                }
                            }, e.get('config.highcharts.account.paymentInterval') || 120000);
                        }
                    }
                },
                title: {
                    text: ""
                },
                xAxis: {
                    ordinal: false,
                    type: "datetime",
                    dateTimeLabelFormats: {
                        day: "%m/%d",
                        week: "%m/%d",
                        month: "%b '%y",
                        year: "%Y"
                    }
                },
                yAxis: {
                    title: {
                        text: "Amount"
                    }
                },
                plotLines: [{
                    value: 0,
                    width: 1,
                    color: "#808080"
                }],
                plotOptions: {
                    series: {
                        borderColor: 'none'
                    }
                },
                legend: {
                    enabled: false,
                    itemStyle: {},
                },
                tooltip: {
                    formatter: function() {
                        return "<b>" + this.y.toFixed(4) + "&nbsp;" + e.get('config.Unit') + "</b><br />" +
                            "<b>" + this.x.toISOString().slice(0, 10) + "</b>";
                    },
                    useHTML: true
                },
                exporting: {
                    enabled: false
                },
                series: [{
                    color: "#E99002",
                    name: "Payments",
                    data: function() {
                        var a = [];
                        if (null != t) {
                            t.forEach(function(d) {
                                var r = new Date(1000 * d.x);
                                var l = r.toLocaleString();
                                var n = d.amount / 1000000000;
                                a.push({x: r, d: l, y: n});
                            });
                        }
                        var now = new Date();
                        var l = now.toLocaleString();
                        var last = {x: now, d: l, y: 0};
                        var interval = e.get('config.highcharts.account.interval') || 120000;
                        if (a.length > 0 && now - a[a.length - 1].x > interval) {
                            a.push(last);
                        }
                        return a;
                    }()
                }]
            };
            a.chart.backgroundColor = this.getWithDefault('config.highcharts.account.backgroundColor', "transparent");
            a.series[0].color = this.getWithDefault('config.highcharts.account.color', ['#E99002'])[0];
            a.xAxis.lineColor = this.getWithDefault('config.highcharts.account.lineColor', "#ccd6eb");
            a.yAxis.lineColor = this.getWithDefault('config.highcharts.account.lineColor', "#ccd6eb");
            a.xAxis.tickColor = this.getWithDefault('config.highcharts.account.tickColor', "#ccd6eb");
            a.yAxis.tickColor = this.getWithDefault('config.highcharts.account.tickColor', "#ccd6eb");
            a.xAxis.gridLineColor = this.getWithDefault('config.highcharts.account.gridLineColor', "#e6e6e6");
            a.yAxis.gridLineColor = this.getWithDefault('config.highcharts.account.gridLineColor', "#e6e6e6");
            a.xAxis.gridLineWidth = this.getWithDefault('config.highcharts.account.gridLineWidthX', "0");
            a.yAxis.gridLineWidth = this.getWithDefault('config.highcharts.account.gridLineWidthY', "1");
            a.legend.itemStyle.color = this.getWithDefault('config.highcharts.account.labelColor', "#fff");
        return a;
    }
})
});
