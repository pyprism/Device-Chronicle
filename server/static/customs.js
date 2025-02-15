function webSocket() {
    const clientID = window.clientID;
    const host = window.location.hostname;
    const port = window.location.port;
    const ws = new WebSocket(`ws://${host}:${port}/analytics_ws/${clientID}`);

    const chartDom = document.getElementById('chart');
    const myChart = echarts.init(chartDom);

    const option = {
        // title: {
        //     text: 'System Analytics',
        //     textStyle: {
        //         fontSize: 18,
        //         fontWeight: 'bold'
        //     },
        //     padding: [10, 0, 0, 0]
        // },
        tooltip: {
            trigger: 'axis'
        },
        legend: {
            data: ['Average Chipset Temp', 'CPU Temp', 'Free RAM', 'Used RAM', 'Used RAM Percentage', 'Packets Received', 'Packets Sent'],
            selected: {
                'Packets Received': false,
                'Packets Sent': false
            }
        },
        xAxis: {
            type: 'category',
            boundaryGap: false,
            data: []
        },
        yAxis: {
            type: 'value',
        },
        series: [
            { name: 'Average Chipset Temp', type: 'line', data: [] },
            { name: 'CPU Temp', type: 'line', data: [] },
            { name: 'Free RAM', type: 'line', data: [] },
            { name: 'Used RAM', type: 'line', data: [] },
            { name: 'Used RAM Percentage', type: 'line', data: [] },
            { name: 'Packets Received', type: 'line', data: [] },
            { name: 'Packets Sent', type: 'line', data: [] }
        ]
    };

    myChart.setOption(option);

// Keep track of the current legend selection and update it when the legend is clicked
    let currentLegend = myChart.getOption().legend[0].selected;
    myChart.on('legendselectchanged', function(params) {
        currentLegend = { ...params.selected };
    });

    ws.onmessage = function(event) {
        const data = JSON.parse(event.data);
        console.log(data);
        const time = new Date().toLocaleTimeString();

        option.xAxis.data.push(time);
        if (option.xAxis.data.length > 500) {
            option.xAxis.data.shift();
        }

        option.series[0].data.push({ value: parseFloat(data.average_chipset_temp), unit: data.average_chipset_temp});
        option.series[1].data.push({ value: parseFloat(data.cpu_temp), unit: data.cpu_temp});
        option.series[2].data.push({ value: parseFloat(data.free_ram), unit: data.free_ram});
        option.series[3].data.push({ value: parseFloat(data.used_ram), unit: data.used_ram});
        option.series[4].data.push({ value: parseFloat(data.used_ram_percentage), unit: data.used_ram_percentage});
        option.series[5].data.push({ value: parseFloat(data.packets_receive), unit: data.packets_receive});
        option.series[6].data.push({ value: parseFloat(data.packets_sent), unit: data.packets_sent});

        option.series.forEach(series => {
            if (series.data.length > 500) {
                series.data.shift();
            }
        });

        option.tooltip = {
            trigger: 'axis',
            formatter: function(params) {
                let result = params[0].axisValue + '<br/>';
                params.forEach(item => {
                    result += item.marker + item.seriesName + ': ' + item.data.unit + '<br/>';
                });
                return result;
            }
        };
        option.legend.selected = currentLegend;
        myChart.setOption(option);
    };

    ws.onopen = function() {
        console.log("Connected to WebSocket");
    };

    ws.onclose = function() {
        console.log("Connection lost, retrying...");
        setTimeout(webSocket, 3000);
    };

    ws.onerror = function(error) {
        console.log("WebSocket Error:", error);
        ws.close();
    };

    window.addEventListener('resize', function() {
        myChart.resize();
    });

}

function setChartDimensions() {
    console.log("Setting chart dimensions");
    const chartDiv = document.getElementById('chart');
    const screenHeight = window.innerHeight;
    const screenWidth = window.innerWidth;
    chartDiv.style.width = (screenWidth * 0.9) + 'px'; // Set width to 90% of the screen width
    chartDiv.style.height = (screenHeight * 0.9) + 'px'; // Set height to 70% of the screen height
}

window.addEventListener('resize', setChartDimensions);


document.addEventListener('DOMContentLoaded', function() {
    setChartDimensions();
    webSocket();
});