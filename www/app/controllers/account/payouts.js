import Ember from 'ember';

export default Ember.Controller.extend({
  applicationController: Ember.inject.controller('application'),
  config: Ember.computed.reads('applicationController.config'),
  stats: Ember.computed.reads('applicationController.model.stats'),
  intl: Ember.inject.service(),

  chartPaymentText: Ember.computed('model', {
    get() {
      var outText = this.get('model.paymentCharts');
      if (!outText) {
        return 0;
      }
      return outText;
    }
  }),

  chartPayment: Ember.computed('intl', 'model.paymentCharts', {
    get() {
        var e = this,
            t = e.getWithDefault("model.paymentCharts"),
            a = {
                chart: {
                    backgroundColor: "rgba(255, 255, 255, 0.1)",
                    type: "column",
                    marginRight: 10,
                    height: 200,
                    events: {
                        load: function() {
                            var self = this;
                            setInterval(function() {
                                if (!self.series) {
                                    return; // FIXME
                                }
                                t = e.getWithDefault("model.paymentCharts");
                                var data = [];
                                t.forEach(function(d) {
                                    var r = new Date(1000 * d.x);
                                    var l = r.toLocaleString();
                                    var n = e.amount / 1000000000;
                                    data.push({x: r, d: l, y: n});
                                });
                                self.series[0].setData(data, true, {}, true);
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
                        day: "%e. %b",
                        week: "%e. %b",
                        month: "%b '%y",
                        year: "%Y"
                    }
                },
                yAxis: {
                    title: {
                        text: "Payment by Account"
                    }
                },
                plotLines: [{
                    value: 0,
                    width: 1,
                    color: "#808080"
                }],
                legend: {
                    enabled: true
                },
                tooltip: {
                    formatter: function() {
                        return "<b>" + Highcharts.dateFormat('%Y-%m-%d', new Date(this.x)) + "<b><br>Payment&nbsp;<b>" + this.y.toFixed(4) + "&nbsp;" + e.get('config.Unit') + "</b>";
                    },
                    useHTML: true
                },
                exporting: {
                    enabled: false
                },
                series: [{
                    color: "#E99002",
                    name: "Payment Series",
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
        return a;
    }
})
});
