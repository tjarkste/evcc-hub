import { reactive } from "vue";
import type { State } from "./types/evcc";
import { ConnectionState } from "./types/evcc";
import { convertToUiLoadpoints } from "./uiLoadpoints";
import { useDebouncedComputed } from "./utils/useDebouncedComputed";
import settings from "./settings";

function setProperty(obj: object, props: string[], value: any) {
  const prop = props.shift();

  if (!props.length) {
    // Leaf node — set or merge the final value.
    //
    // Leaf guard: never overwrite an existing array or plain-object with a
    // primitive. evcc publishes count topics like site/pv=1 and
    // site/loadpoints=1 that must not corrupt the pv[] / loadpoints[] arrays.
    // @ts-expect-error no-explicit-any
    const existing = obj[prop];
    const incomingIsPrimitive = value === null || typeof value !== "object";
    const existingIsContainer =
      existing !== null && existing !== undefined && typeof existing === "object";
    if (incomingIsPrimitive && existingIsContainer) {
      return; // preserve the array / object
    }

    if (value && typeof value === "object" && !Array.isArray(value)) {
      // Merge plain objects; replace if existing is an array or absent.
      if (existingIsContainer && !Array.isArray(existing)) {
        // @ts-expect-error no-explicit-any
        obj[prop] = { ...existing, ...value };
      } else {
        // @ts-expect-error no-explicit-any
        obj[prop] = value;
      }
    } else {
      // @ts-expect-error no-explicit-any
      obj[prop] = value;
    }
    return;
  }

  // Intermediate node — must recurse deeper.
  //
  // Recursion guard: if the slot is missing or holds a primitive (e.g. a count
  // topic like site/pv=1 wrote a number here before array data arrived),
  // replace it with the correct container type before recursing.
  // @ts-expect-error no-explicit-any
  if (!obj[prop] || typeof obj[prop] !== "object") {
    const nextKey = props[0];
    // @ts-expect-error no-explicit-any
    obj[prop] = /^\d+$/.test(nextKey) ? [] : {};
  }

  // @ts-expect-error no-explicit-any
  setProperty(obj[prop], props, value);
}

const initialState: State = {
  offline: false,
  connectionState: ConnectionState.OFFLINE,
  lastDataAt: null,
  loadpoints: [],
  vehicles: {},
  forecast: {},
};

const state = reactive(initialState);

// create derived loadpoints array with ui specific fields (defaults, browser settings, ...); debounce for better performance
const uiLoadpoints = useDebouncedComputed(
  () => convertToUiLoadpoints(state.loadpoints, state.vehicles),
  () => [state.loadpoints, state.vehicles, settings.loadpoints],
  50
);

export interface Store {
  state: State; // raw state from websocket
  uiLoadpoints: typeof uiLoadpoints;
  offline(value: boolean): void;
  update(msg: any): void;
  reset(): void;
}

const store: Store = {
  state,
  uiLoadpoints,
  offline(value: boolean) {
    state.offline = value;
  },
  update(msg) {
    Object.keys(msg).forEach(function (k) {
      if (k === "log") {
        window.app.raise(msg[k]);
      } else {
        setProperty(state, k.split("."), msg[k]);
      }
    });
  },
  reset() {
    console.log("resetting state");
    // reset to initial state
    Object.keys(initialState).forEach(function (k) {
      if (k === "offline" || k === "connectionState" || k === "lastDataAt") return;

      // @ts-expect-error no-explicit-any
      if (Array.isArray(initialState[k])) {
        // @ts-expect-error no-explicit-any
        state[k] = [];
      } else {
        // @ts-expect-error no-explicit-any
        state[k] = undefined;
      }
    });
  },
};

export default store;
