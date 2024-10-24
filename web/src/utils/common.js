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
    const value = field.split('.').reduce((obj, key) => (obj && obj[key] !== undefined) ? obj[key] : undefined, row);
    if(field == "spec.variables") {
        return JSON.stringify(value, null, 4);
    } else if(field == "spec.tasks") {
        return JSON.stringify(value, null, 4);
    }
    return value;
}

export { proxyVariablesToJsonObject, formatObject };
