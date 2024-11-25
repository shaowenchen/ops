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

  if (field === "spec.variables" || field === "spec.tasks") {
    return JSON.stringify(value, null, 4);
  } else if (field === "status.heartTime") {
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
  }

  return value;
}

export { proxyVariablesToJsonObject, formatObject };
