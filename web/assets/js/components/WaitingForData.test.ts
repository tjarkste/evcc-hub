import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import WaitingForData from "./WaitingForData.vue";
import store from "../store";
import { ConnectionState } from "../types/evcc";

const mockT = (key: string) => key;

describe("WaitingForData", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    store.state.connectionState = ConnectionState.OFFLINE;
    store.state.lastDataAt = null;
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("shows connecting key for RECONNECTING state", () => {
    store.state.connectionState = ConnectionState.RECONNECTING;
    const wrapper = mount(WaitingForData, { global: { mocks: { $t: mockT } } });
    expect(wrapper.text()).toContain("hub.waiting.connecting");
  });

  it("shows connecting key for OFFLINE state", () => {
    store.state.connectionState = ConnectionState.OFFLINE;
    const wrapper = mount(WaitingForData, { global: { mocks: { $t: mockT } } });
    expect(wrapper.text()).toContain("hub.waiting.connecting");
  });

  it("shows waitingForData key when CONNECTED", () => {
    store.state.connectionState = ConnectionState.CONNECTED;
    const wrapper = mount(WaitingForData, { global: { mocks: { $t: mockT } } });
    expect(wrapper.text()).toContain("hub.waiting.waitingForData");
  });
});
