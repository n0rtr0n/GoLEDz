let ws;
const pixelVisibility = 255

const canvas = document.getElementById("canvas");
const $ctx = canvas.getContext("2d");
$ctx.fillStyle = "#000000";
$ctx.fillRect(0, 0, canvas.width, canvas.height);

function connectWebSocket() {
    ws = new WebSocket('ws://127.0.0.1:8008/socket');

    ws.onopen = function() {
        console.log('WebSocket connection established.');
    };

    ws.onclose = function(event) {
        console.log('WebSocket connection closed.');

        // Reconnect only if the connection is closed, not if it's in the process of closing.
        if (event.code !== 1000) {
            console.log('Reconnecting...');
            setTimeout(connectWebSocket, 5000); // Retry after 5 seconds
        }
    };

    ws.onerror = function(error) {
        console.error('WebSocket error:', error);
    };

    ws.onmessage = function(event) {
      const message = event.data;
      processMessage(message);
    };
}

connectWebSocket();

const processMessage = (event) => {
  const dataFromServer = JSON.parse(event);

  $ctx.fillStyle = '#000000';
  $ctx.fillRect(0, 0, canvas.width, canvas.height);
  // Oddly enough, this is pretty quick in Firefox but is terribly slow in Chrome?
  // TODO: switch fill method based on which browser is detected?
  let $px = $ctx.createImageData(3, 3);

  dataFromServer.pixels.map(pixel => {
    for (let i = 0; i < $px.data.length; i += 4) {
      $px.data[i + 0] = pixel.r;  //red
      $px.data[i + 1] = pixel.g;  //green
      $px.data[i + 2] = pixel.b;  //blue
      $px.data[i + 3] = pixelVisibility;
    }
    $ctx.putImageData($px, pixel.x, canvas.height - pixel.y);
    return true;
  });
}