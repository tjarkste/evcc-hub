import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import WaitingForData from "./WaitingForData.vue";
import store from "../store";
import { ConnectionState } from "../types/evcc";

describe("WaitingForData", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    store.state.connectionState = ConnectionState.OFFLINE;
    store.state.lastDataAt = null;
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("shows same message for RECONNECTING as for OFFLINE", () => {
    store.state.connectionState = ConnectionState.RECONNECTING;
    const wrapper = mount(WaitingForData);
    expect(wrapper.text()).toContain("Verbinde mit MQTT-Broker...");
  });

  it("shows connecting message for OFFLINE state", () => {
    store.state.connectionState = ConnectionState.OFFLINE;
    const wrapper = mount(WaitingForData);
    expect(wrapper.text()).toContain("Verbinde mit MQTT-Broker...");
  });

  it("shows waiting message when CONNECTED", () => {
    store.state.connectionState = ConnectionState.CONNECTED;
    const wrapper = mount(WaitingForData);
    expect(wrapper.text()).toContain("Verbunden — warte auf Daten");
  });
});
