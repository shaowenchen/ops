function formatMemory(memory) {
  const units = ["Ki", "Mi", "Gi", "Ti"];
  let value = parseFloat(memory);
  let unitIndex = 0;

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex++;
  }

  return `${value.toFixed(2)} ${units[unitIndex]}`;
}

export { formatMemory };
