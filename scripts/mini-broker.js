// Minimaler MQTT-over-WebSocket Broker für lokale Entwicklung
// Nutzt nur mqtt-packet und ws (beide in web/node_modules)

const mqttPacket = require('../web/node_modules/mqtt-packet');

let wss;
try {
  wss = new (require('../web/node_modules/ws').Server)({ port: 9001 });
} catch(e) {
  console.error('ws not found:', e.message);
  process.exit(1);
}

const clients = new Map(); // clientId -> { socket, subscriptions, protocolVersion }
const retained = new Map(); // topic -> payload

function matchTopic(filter, topic) {
  if (filter === topic) return true;
  const fp = filter.split('/');
  const tp = topic.split('/');
  for (let i = 0; i < fp.length; i++) {
    if (fp[i] === '#') return true;
    if (fp[i] !== '+' && fp[i] !== tp[i]) return false;
  }
  return fp.length === tp.length;
}

function sendPacket(socket, packet, protocolVersion) {
  try {
    const buf = mqttPacket.generate({ ...packet, protocolVersion });
    socket.send(buf);
  } catch(e) {
    console.error('sendPacket error:', e.message, JSON.stringify(packet));
  }
}

wss.on('connection', (socket) => {
  let clientId = null;
  let protocolVersion = 4;
  // Try v5 parser first, fall back to v4 on error
  let parser = mqttPacket.parser({ protocolVersion: 5 });

  parser.on('packet', (packet) => {
    if (packet.cmd === 'connect') {
      protocolVersion = packet.protocolVersion || 4;
      clientId = packet.clientId || ('client_' + Date.now());
      clients.set(clientId, { socket, subscriptions: [], protocolVersion });

      const connack = protocolVersion >= 5
        ? { cmd: 'connack', sessionPresent: false, reasonCode: 0 }
        : { cmd: 'connack', sessionPresent: false, returnCode: 0 };
      sendPacket(socket, connack, protocolVersion);
      console.log(`[+] ${clientId} connected (v${protocolVersion})`);
    }

    else if (packet.cmd === 'subscribe') {
      const client = clients.get(clientId);
      if (!client) return;
      packet.subscriptions.forEach(s => {
        client.subscriptions.push(s.topic);
        // Send retained messages
        for (const [topic, payload] of retained) {
          if (matchTopic(s.topic, topic)) {
            const pub = { cmd: 'publish', topic, payload, qos: 0, retain: true, dup: false };
            sendPacket(socket, pub, protocolVersion);
          }
        }
      });
      const suback = protocolVersion >= 5
        ? { cmd: 'suback', messageId: packet.messageId, properties: {}, reasonCodes: packet.subscriptions.map(() => 0) }
        : { cmd: 'suback', messageId: packet.messageId, granted: packet.subscriptions.map(s => s.qos) };
      sendPacket(socket, suback, protocolVersion);
    }

    else if (packet.cmd === 'publish') {
      if (packet.retain) retained.set(packet.topic, packet.payload);
      for (const [, client] of clients) {
        for (const filter of client.subscriptions) {
          if (matchTopic(filter, packet.topic)) {
            const pub = { cmd: 'publish', topic: packet.topic, payload: packet.payload, qos: 0, retain: false, dup: false };
            sendPacket(client.socket, pub, client.protocolVersion);
            break;
          }
        }
      }
    }

    else if (packet.cmd === 'pingreq') {
      sendPacket(socket, { cmd: 'pingresp' }, protocolVersion);
    }

    else if (packet.cmd === 'disconnect') {
      socket.close();
    }
  });

  socket.on('message', (data) => {
    parser.parse(data);
  });

  socket.on('close', () => {
    if (clientId) {
      clients.delete(clientId);
      console.log(`[-] ${clientId} disconnected`);
    }
  });

  parser.on('error', () => {});
});

console.log('Mini MQTT broker running on ws://localhost:9001');
