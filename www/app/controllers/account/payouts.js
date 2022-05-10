import Ember from 'ember';

export default Ember.Controller.extend({
  applicationController: Ember.inject.controller('application'),
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
                    backgroundColor: "rgba(0, 0, 0, 0.1)",
                    type: "column",
                    marginRight: 10,
                    height: 200,
                    events: {
                        load: function() {
                            var series = this.series[0];
                            setInterval(function() {
                                var x = (new Date()).getDate(),
                                    y = e.getWithDefault("model.paymentCharts");
                                series.addPoint([x, y], true, true);
                            }, 1090000000);
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
                    enabled: true,
                    itemStyle:
                      {
                        color:"#ccc"
                      },
                },
                tooltip: {
                    formatter: function() {
                        return "<b>" + HighCharts.dateFormat('%Y-%m-%d', new Date(this.x)) + "<b><br>Payment&nbsp;<b>" + this.y.toFixed(4) + "&nbsp;CLO</b>";
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
                        var e, a = [];
                        if (null != t) {
                            for (e = 0; e <= t.length - 1; e += 1) {
                                var n = 0,
                                    r = 0,
                                    l = 0;
                                    r = new Date(1e3 * t[e].x);
                                    l = r.toLocaleString();
                                    n = t[e].amount / 1000000000;
                                    a.push({
                                        x: r,
                                        d: l,
                                        y: n
                                    });
                            }
                        } else {
                            a.push({
                                x: 0,
                                d: 0,
                                y: 0
                           });
                        }
                        return a;
                    }()
                }]
            };
        return a;
    }
})
});
