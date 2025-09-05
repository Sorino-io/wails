import { createI18n } from 'vue-i18n'
import ar from './locales/ar.json'
import en from './locales/en.json'

const i18n = createI18n({
  locale: 'ar', // default locale
  fallbackLocale: 'en',
  messages: {
    ar,
    en
  },
  legacy: false,
  globalInjection: true
})

export default i18n
