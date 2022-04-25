var xmlhttp = new XMLHttpRequest();
var url = "http://localhost:10000/temps"//
xmlhttp.open('GET', url, true);
xmlhttp.send();

xmlhttp.onreadystatechange = function() {
  if(this.readyState == 4 && this.status == 200){
    console.log(this.responseText)
    var data = JSON.parse(this.responseText);
  
    var timestamp = data.map(function(elem) {
      return elem.timestamp;
    });
    var deviceA = data.filter(e=> {
      return e.deviceid == "deviceA";
    }).map(function(elem) {
      return elem.temperature;
    });
    var deviceB = data.filter(e=> {
      return e.deviceid == "deviceB";
    }).map(function(elem) {
      return elem.temperature;
    });

    // var temperature = data.map(function(elem) {
    //   return elem.temperature;
    // });
    
    var ctxL = document.getElementById('canvas').getContext('2d');
    var mychart = new Chart(ctxL, {
    type: 'line',
    data: {
      labels: timestamp,
      datasets: [{
          label: 'Device A',
          data: deviceA,//
          backgroundColor: [
            'rgba(105, 0, 132, .2)',
          ],
          borderColor: [
            'rgba(200, 99, 132, .7)',
          ],
          borderWidth: 2
        },
        {
          label: 'Device B',
          data: deviceB,//
          backgroundColor: [
            'rgba(0, 137, 132, .2)',
          ],
          borderColor: [
            'rgba(0, 10, 130, .7)',
          ],
          borderWidth: 2
        }
      ]
    },
    options: {
      responsive: true
    }
  });
  }
}
