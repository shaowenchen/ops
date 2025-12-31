function proxyVariablesToJsonObject(proxy) {
  const jsonObject = {};
  Object.keys(proxy).forEach((key) => {
    if (proxy[key].value != undefined) {
      jsonObject[key] = proxy[key].value;
    } else if (proxy[key].default != undefined) {
      jsonObject[key] = proxy[key].default;
    }
  });
  return jsonObject;
}

function formatObject(row, field) {
  const value = field
    .split(".")
    .reduce(
      (obj, key) => (obj && obj[key] !== undefined ? obj[key] : undefined),
      row
    );

  if (field === "spec.tasks") {
    if (Array.isArray(value)) {
      const taskRefs = value
        .map((item) => item.taskRef)
        .filter(Boolean)
        .join(", ");
      return taskRefs || "No taskRef found";
    }
  } else if (field === "status.heartTime" || field === "status.startTime" || field === "event.time") {
    // For event.time, prefer the formatted time field from backend (local timezone)
    if (field === "event.time" && row.time) {
      return row.time;
    }
    const date = new Date(value);
    if (!isNaN(date)) {
      const options = {
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        timeZone: "UTC",
        hour12: false,
      };

      const formatter = new Intl.DateTimeFormat("en-US", options);
      return formatter.format(date) + "Z";
    } else {
      return "Invalid Date";
    }
  } else if (field === "time") {
    // Direct time field from EventData (already formatted in local timezone)
    return value || "";
  } else if (field == "spec.variables") {
    if (typeof value === "object" && value !== null) {
      const keys = Object.keys(value).join(", ");
      return keys;
    }
  }
  return value;
}

export { proxyVariablesToJsonObject, formatObject };
