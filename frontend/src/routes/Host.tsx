import { useEffect, useRef, useState } from "react";
import { WsClient } from "../ws/client";
import type { OutgoingMessage } from "../ws/messages";

function getHostWsUrl(): string {
  const params = new URLSearchParams(window.location.search);
  const token = params.get("token") ?? "";
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/ws?token=${token}`;
}

export default function Host(): JSX.Element {
  const clientRef = useRef<WsClient | null>(null);
  const [connected, setConnected] = useState(false);
  const [lastEvent, setLastEvent] = useState<string | null>(null);

  useEffect(() => {
    const client = new WsClient(getHostWsUrl());
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
      <h1>Quizmaster Panel</h1>
      <p>Status: {connected ? "connected" : "disconnected"}</p>
      {lastEvent !== null && <p>Last event: {lastEvent}</p>}
    </div>
  );
}
