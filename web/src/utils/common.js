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

export { proxyVariablesToJsonObject };
