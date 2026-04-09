import { useEffect, useRef, useState } from "react";
import { WsClient } from "../ws/client";
import type { OutgoingMessage } from "../ws/messages";

function getPlayWsUrl(): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/ws?room=play`;
}

export default function Play(): JSX.Element {
  const clientRef = useRef<WsClient | null>(null);
  const [connected, setConnected] = useState(false);
  const [lastEvent, setLastEvent] = useState<string | null>(null);

  useEffect(() => {
    const client = new WsClient(getPlayWsUrl());
    clientRef.current = client;

    client.onMessage((msg: OutgoingMessage) => {
      setLastEvent(msg.event);
    });

    client.on("reconnect_failed", () => {
      setConnected(false);
    });

    client.connect();
    setConnected(true);

    return () => {
      client.disconnect();
    };
  }, []);

  return (
    <div>
      <h1>Play</h1>
      <p>Status: {connected ? "connected" : "disconnected"}</p>
      {lastEvent !== null && <p>Last event: {lastEvent}</p>}
    </div>
  );
}
