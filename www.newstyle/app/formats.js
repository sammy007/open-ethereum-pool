var hhmmss = {
  hour:   'numeric',
  minute: 'numeric',
  second: 'numeric'
};

export default {
  time: {
    hhmmss: hhmmss
  },
  date: {
    hhmmss: hhmmss
  },
  number: {
    EUR: { style: 'currency', currency: 'EUR', minimumFractionDigits: 2, maximumFractionDigits: 2 },
    USD: { style: 'currency', currency: 'USD', minimumFractionDigits: 2, maximumFractionDigits: 2 }
  }
};
