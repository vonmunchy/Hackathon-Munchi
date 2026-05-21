interface CaptionProduct {
  name: string
  basePrice: number
  category?: string
  description?: string
}

interface CaptionStore {
  name: string
  slug: string
  instagramUsername?: string
}

interface CaptionVariant {
  variantName: string
  stockAvailable: number
  priceOverride?: number
}

export function generateInstagramCaption(
  product: CaptionProduct,
  store: CaptionStore,
  variants: CaptionVariant[],
  storefrontUrl: string,
): string {
  const lowStock = variants.some((v) => v.stockAvailable <= 3)
  const variantNames = variants.map((v) => v.variantName).join(', ')

  let caption = `✨ ${product.name} ✨\n\n`
  if (product.description) {
    caption += `${product.description}\n\n`
  }
  caption += `💰 Starting from MVR ${product.basePrice.toFixed(2)}\n`
  if (variantNames) {
    caption += `🎨 Available in: ${variantNames}\n`
  }
  if (lowStock) {
    caption += `🔥 Limited stock! Don't miss out!\n`
  }
  caption += `\n🛒 Shop now at ${storefrontUrl}\n`
  caption += `📲 Pay securely with Swipe\n`
  caption += `\n📍 ${store.name}\n`
  caption += `🔗 Link in bio`

  return caption
}

export function generateFacebookCaption(
  product: CaptionProduct,
  store: CaptionStore,
  variants: CaptionVariant[],
  storefrontUrl: string,
): string {
  const variantNames = variants.map((v) => v.variantName).join(', ')
  const lowStock = variants.some((v) => v.stockAvailable <= 3)

  let caption = `New from ${store.name}!\n\n`
  caption += `${product.name}\n`
  if (product.description) {
    caption += `${product.description}\n`
  }
  caption += `\nPrice: MVR ${product.basePrice.toFixed(2)}\n`
  if (variantNames) {
    caption += `Options: ${variantNames}\n`
  }
  if (lowStock) {
    caption += `\nHurry - limited stock remaining!\n`
  }
  caption += `\nSecure payment via Swipe. Fast delivery across the Maldives.\n`
  caption += `\nShop now: ${storefrontUrl}`

  return caption
}

export function generateHashtags(
  product: CaptionProduct,
  store: CaptionStore,
): string {
  const categoryTag = product.category
    ? `#${product.category.replace(/[^a-zA-Z0-9]/g, '')}`
    : ''
  const storeTag = `#${store.slug.replace(/-/g, '').replace(/[^a-zA-Z0-9]/g, '')}`

  const tags = [
    '#MaldivesShopping',
    categoryTag,
    storeTag,
    '#SwipePay',
    '#MVR',
    '#ShopLocal',
    '#MaldivesStore',
    '#IslandShopping',
    '#OnlineShopping',
    '#MadeInMaldives',
  ].filter(Boolean)

  return tags.join(' ')
}
