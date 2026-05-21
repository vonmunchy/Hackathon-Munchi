---
title: "Converting a Next.js + Convex Web App to Expo React Native for Hackathon Demos"
date: 2026-05-21
category: architecture-patterns
module: swipe-mobile
problem_type: architecture_pattern
component: tooling
severity: medium
applies_when:
  - "Hackathon or demo day requiring native mobile experience"
  - "Existing Next.js + Convex web app needs a mobile companion"
  - "Both iOS and Android support needed without separate codebases"
  - "No Apple/Google developer accounts available for store distribution"
  - "Time constraint under 48 hours for mobile delivery"
tags:
  - expo
  - react-native
  - nextjs-migration
  - convex
  - nativewind
  - expo-router
  - expo-go
  - hackathon
---

# Converting a Next.js + Convex Web App to Expo React Native for Hackathon Demos

## Context

The team had a working Next.js 16 + Convex 1.39 + Tailwind 4 web app (Swipe Social Storefront -- a Maldives fintech marketplace with P2P crypto exchange) for a hackathon demo. Evaluators needed to experience the app natively on their phones via QR code scanning, on both iOS and Android, without any app store accounts or build infrastructure.

Four approaches were evaluated:

| Approach | Time | Native Feel | All Features | iOS+Android | Server Needed |
|----------|------|-------------|-------------|-------------|---------------|
| PWA | 30 min | Low | Yes | Partial | No |
| Capacitor | 2-3 hrs | Medium | Yes | Android only* | No |
| Expo Go (chosen) | 1-2 days | High | Partial | Yes | Yes |
| EAS Build | 2-3 days | High | Partial | Android only* | No |

*iOS requires paid Apple Developer account ($99/yr)

**Decision**: Expo Go for the demo -- evaluators install the free Expo Go app, scan a QR code, and get a native experience on both platforms. The laptop runs the dev server during the demo.

## Guidance

### Stack Selection

- **Expo SDK 54** + **expo-router v6** for file-based routing (mirrors Next.js App Router)
- **NativeWind v4** for Tailwind-like styling (70-80% of classes transfer directly)
- **Convex 1.39** backend shared unchanged -- copy the `convex/` directory directly
- **expo-image** for optimized image rendering
- **react-native-reanimated** for animations (replaces Framer Motion)

### Critical Configuration

**babel.config.js** -- preset order matters:
```js
module.exports = function (api) {
  api.cache(true);
  return {
    presets: [
      ["babel-preset-expo", { jsxImportSource: "nativewind" }],
      "nativewind/babel",
    ],
    plugins: ["react-native-reanimated/plugin"], // must be last
  };
};
```

**metro.config.js** -- NativeWind Metro integration:
```js
const { getDefaultConfig } = require("expo/metro-config");
const { withNativeWind } = require("nativewind/metro");
const config = getDefaultConfig(__dirname);
module.exports = withNativeWind(config, { input: "./global.css" });
```

**ConvexProvider** -- disable web-only feature for React Native:
```tsx
const convex = new ConvexReactClient(
  process.env.EXPO_PUBLIC_CONVEX_URL!,
  { unsavedChangesWarning: false }
);
```

**Environment variables** -- use `EXPO_PUBLIC_` prefix:
```
EXPO_PUBLIC_CONVEX_URL=https://your-deployment.convex.cloud
```

### Dependencies: Use `npx expo install`

Always use `npx expo install` instead of `npm install` to avoid peer dependency conflicts. Two undocumented manual installs are required:

```bash
npx expo install babel-preset-expo react-native-worklets
```

### Route Mapping Pattern

| Next.js App Router | expo-router | Notes |
|-------------------|-------------|-------|
| `app/page.tsx` | `app/(tabs)/index.tsx` | Home tab |
| `app/marketplace/page.tsx` | `app/(tabs)/marketplace.tsx` | Tab screen |
| `app/exchange/page.tsx` | `app/(tabs)/exchange.tsx` | Tab screen |
| `app/shop/[slug]/product/[id]/page.tsx` | `app/product/[id].tsx` | Stack screen |
| `app/checkout/[orderId]/page.tsx` | `app/checkout/[orderId].tsx` | Stack screen |
| `app/seller/login/page.tsx` | `app/seller/login.tsx` | Outside tabs |

### Navigation Architecture

Web app's `BuyerShell` bottom bar (Home / Marketplace / Exchange) maps to a `(tabs)` layout group. Seller flow lives outside tabs as a separate route group accessed via explicit navigation.

```tsx
// app/(tabs)/_layout.tsx
<Tabs>
  <Tabs.Screen name="index" options={{ title: "Home" }} />
  <Tabs.Screen name="marketplace" options={{ title: "Marketplace" }} />
  <Tabs.Screen name="exchange" options={{ title: "Exchange" }} />
</Tabs>
```

### Expo Go Distribution

1. Install ngrok auth: the bundled `@expo/ngrok` (v2.x) uses `authtoken`, not `config add-authtoken`
2. Start with tunnel: `npx expo start --tunnel`
3. Evaluators scan QR code with Expo Go app (free on both iOS and Android)
4. Laptop must remain running during the demo

### Design Matching via Chrome DevTools MCP

The iterative process for achieving visual parity:

1. Run web dev server at `localhost:3000`
2. Run Expo web preview at `localhost:8090` (install `react-native-web`)
3. Use Chrome DevTools MCP with mobile emulation (`390x844x3,mobile,touch`)
4. Screenshot each web page, compare with Expo preview
5. Iterate screens until visual parity

## Why This Matters

- **Time**: Full React Native rewrite takes weeks; this pattern delivers a working native demo in 1-2 days
- **Backend reuse**: Convex hooks (`useQuery`, `useMutation`) work identically in React Native -- zero API changes, zero schema changes
- **Reach**: Expo Go works on both iOS and Android with no build infrastructure or developer accounts
- **Future path**: Same codebase can produce standalone APK/IPA via `eas build` when ready for production

## When to Apply

- Hackathon demo where evaluators need a mobile-native experience
- Existing Next.js + Convex web app that needs a mobile companion quickly
- Team familiar with React/Tailwind but not native mobile development
- Both iOS and Android support required without app store accounts
- Time constraint under 48 hours

Do **not** apply when:
- Full production mobile app is needed (consider a proper React Native project)
- Offline support is required (Expo Go needs network connectivity)
- App store distribution is required immediately (use EAS Build instead)

## Examples

### Component Translation

| Web (Next.js + Tailwind) | Mobile (Expo + NativeWind) |
|--------------------------|---------------------------|
| `<div className="...">` | `<View className="...">` |
| `<p>`, `<span>`, `<h1>` | `<Text className="...">` |
| `<img src="...">` | `<Image source={{ uri: "..." }}>` (expo-image) |
| `<input>` | `<TextInput>` |
| `<Link href="/x">` | `<Link href="/x">` (expo-router) |
| `useRouter().push()` | `router.push()` (expo-router) |
| `window.alert()` | `Alert.alert()` |
| CSS Grid | Flex with `flex-wrap` (no CSS Grid in RN) |
| `hover:` states | `active:` or `onPressIn` |
| `localStorage` | `AsyncStorage` or `SecureStore` |

### Android Alert.prompt Gotcha

`Alert.prompt` is iOS-only. Use a custom Modal:

```tsx
// Works on both iOS and Android
<Modal visible={showInput} transparent animationType="slide">
  <View className="flex-1 justify-end bg-black/40">
    <View className="bg-white rounded-t-3xl p-5">
      <TextInput
        value={inputValue}
        onChangeText={setInputValue}
        className="border border-slate-200 rounded-xl px-4 py-3"
      />
      <TouchableOpacity onPress={() => handleSubmit(inputValue)}>
        <Text>Continue</Text>
      </TouchableOpacity>
    </View>
  </View>
</Modal>
```

### Convex Provider Setup

```tsx
// Next.js (web)
const convex = new ConvexReactClient(process.env.NEXT_PUBLIC_CONVEX_URL!);

// Expo (mobile) -- note the prefix and option differences
const convex = new ConvexReactClient(
  process.env.EXPO_PUBLIC_CONVEX_URL!,
  { unsavedChangesWarning: false }
);
```

### Final Output

- 12 screens with visual parity to the web app
- 3 shared components (ProductCard, MvrAmount, LoadingSpinner)
- Android bundle: 4.51 MB, zero TypeScript errors
- Convex backend shared with zero modifications

## Related

- [Dual Layout Shell Pattern](../dual-layout-shell-js-viewport-detection-2026-05-20.md) -- the web-based mobile/desktop shell this native app supersedes for mobile users
- [Swipe/Next.js/Convex Split Responsibility](../swipe-nextjs-convex-split-responsibility-2026-05-20.md) -- Convex backend architecture reused by the Expo app
- [Multi-tenant Session Auth](../../../swipe-social-storefront/docs/solutions/architecture-patterns/multi-tenant-session-auth-convex-2026-05-21.md) -- session auth pattern the Expo app integrates with
- Original requirements scoped "No native mobile app (responsive web only)" -- this work supersedes that boundary
