import { useState } from "react";
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  TextInput,
  Modal,
  Alert,
  KeyboardAvoidingView,
  Platform,
  Linking,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { Image } from "expo-image";
import { useRouter, useLocalSearchParams } from "expo-router";
import { useQuery, useMutation } from "convex/react";
import { api } from "../../convex/_generated/api";
import type { Id } from "../../convex/_generated/dataModel";
import LoadingSpinner from "../../components/LoadingSpinner";
import { Ionicons } from "@expo/vector-icons";

const LOCATIONS = ["Male'", "Hulhumale'"] as const;

export default function ProductDetail() {
  const router = useRouter();
  const { id } = useLocalSearchParams<{ id: string }>();
  const product = useQuery(
    api.products.getWithVariants,
    id ? { productId: id as Id<"products"> } : "skip"
  );
  const store = useQuery(
    api.stores.getById,
    product?.storeId ? { storeId: product.storeId } : "skip"
  );
  const createOrder = useMutation(api.orders.createOrder);

  const [selectedVariantIdx, setSelectedVariantIdx] = useState(0);
  const [quantity, setQuantity] = useState(1);
  const [showForm, setShowForm] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  // Delivery form state
  const [customerName, setCustomerName] = useState("");
  const [customerPhone, setCustomerPhone] = useState("");
  const [deliveryLocation, setDeliveryLocation] = useState<string>(
    LOCATIONS[0]
  );
  const [deliveryAddress, setDeliveryAddress] = useState("");

  if (product === undefined) {
    return (
      <SafeAreaView className="flex-1 bg-white">
        <LoadingSpinner message="Loading product..." />
      </SafeAreaView>
    );
  }

  if (product === null) {
    return (
      <SafeAreaView className="flex-1 bg-white items-center justify-center px-8">
        <Ionicons name="bag-remove-outline" size={48} color="#cbd5e1" />
        <Text className="text-lg font-semibold text-slate-700 mt-4">
          Product not found
        </Text>
        <TouchableOpacity onPress={() => router.back()} className="mt-4">
          <Text className="text-[#9333ea] font-medium">Go back</Text>
        </TouchableOpacity>
      </SafeAreaView>
    );
  }

  const variants = product.variants ?? [];
  const selectedVariant = variants[selectedVariantIdx];
  const unitPrice = selectedVariant?.priceOverride ?? product.basePrice;
  const stock = selectedVariant?.stockAvailable ?? 0;
  const totalPrice = unitPrice * quantity;

  const openWhatsApp = () => {
    const phone = store?.whatsappNumber;
    if (!phone) {
      Alert.alert("No WhatsApp", "This seller has not set a WhatsApp number.");
      return;
    }
    const cleaned = phone.replace(/\D/g, "");
    const url = `https://wa.me/${cleaned}?text=${encodeURIComponent(
      `Hi! I'm interested in "${product.name}".`
    )}`;
    Linking.openURL(url);
  };

  const handleSubmitOrder = async () => {
    if (!customerName.trim()) {
      Alert.alert("Missing Info", "Please enter your name.");
      return;
    }
    if (!/^\d{7}$/.test(customerPhone)) {
      Alert.alert("Invalid Phone", "Phone number must be exactly 7 digits.");
      return;
    }
    if (!deliveryAddress.trim()) {
      Alert.alert("Missing Info", "Please enter your delivery address.");
      return;
    }
    if (!selectedVariant) return;

    setSubmitting(true);
    try {
      const result = await createOrder({
        storeId: product.storeId,
        productId: product._id,
        variantId: selectedVariant._id,
        quantity,
        customerName: customerName.trim(),
        customerPhone,
        deliveryLocation,
        deliveryAddress: deliveryAddress.trim(),
      });

      if (result.success) {
        setShowForm(false);
        router.push(`/checkout/${result.orderId}?token=${result.accessToken}`);
      } else {
        Alert.alert("Order Failed", result.reason ?? "Something went wrong.");
      }
    } catch (err: any) {
      Alert.alert("Error", err.message ?? "Could not create order.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <SafeAreaView className="flex-1 bg-white" edges={["top"]}>
      {/* Top bar */}
      <View className="flex-row items-center px-4 py-3 border-b border-slate-100">
        <TouchableOpacity
          onPress={() => router.back()}
          className="mr-3 w-9 h-9 rounded-full bg-slate-100 items-center justify-center"
        >
          <Ionicons name="arrow-back" size={20} color="#1e293b" />
        </TouchableOpacity>
        <Text
          className="text-base font-semibold text-slate-900 flex-1"
          numberOfLines={1}
        >
          Swipe Store
        </Text>
      </View>

      <ScrollView className="flex-1" showsVerticalScrollIndicator={false}>
        {/* Product image */}
        <Image
          source={{ uri: product.imageUrls?.[0] ?? "" }}
          className="w-full aspect-square"
          contentFit="cover"
          transition={200}
        />

        <View className="px-5 pt-5 pb-8">
          {/* Product name */}
          <Text className="text-2xl font-bold text-slate-900">
            {product.name}
          </Text>

          {/* Price */}
          <Text className="text-xl font-bold text-[#9333ea] mt-1">
            MVR {unitPrice.toFixed(2)}
          </Text>

          {/* Description */}
          {product.description ? (
            <Text className="text-sm text-slate-500 mt-3 leading-5">
              {product.description}
            </Text>
          ) : null}

          {/* Variant selector */}
          {variants.length > 0 ? (
            <View className="mt-6">
              <Text className="text-sm font-semibold text-slate-700 mb-2.5">
                Option
              </Text>
              <View className="flex-row flex-wrap gap-2">
                {variants.map((v, idx) => {
                  const outOfStock = (v.stockAvailable ?? 0) <= 0;
                  const isSelected = selectedVariantIdx === idx;

                  return (
                    <TouchableOpacity
                      key={v._id}
                      onPress={() => {
                        if (!outOfStock) {
                          setSelectedVariantIdx(idx);
                          setQuantity(1);
                        }
                      }}
                      disabled={outOfStock}
                      className={`px-4 py-2 rounded-full border ${
                        outOfStock
                          ? "bg-slate-50 border-slate-200"
                          : isSelected
                            ? "bg-[#9333ea] border-[#9333ea]"
                            : "bg-white border-slate-200"
                      }`}
                    >
                      <Text
                        className={`text-sm font-medium ${
                          outOfStock
                            ? "text-slate-400 line-through"
                            : isSelected
                              ? "text-white"
                              : "text-slate-700"
                        }`}
                      >
                        {outOfStock
                          ? `${v.variantName} Out`
                          : v.variantName}
                      </Text>
                    </TouchableOpacity>
                  );
                })}
              </View>
            </View>
          ) : null}

          {/* Stock count */}
          {selectedVariant ? (
            <Text
              className={`text-xs mt-3 ${
                stock <= 3 && stock > 0
                  ? "text-red-500 font-semibold"
                  : stock === 0
                    ? "text-red-500 font-semibold"
                    : "text-slate-400"
              }`}
            >
              {stock > 0 ? `${stock} available` : "Out of stock"}
            </Text>
          ) : null}

          {/* Quantity selector */}
          <View className="mt-6">
            <Text className="text-sm font-semibold text-slate-700 mb-2.5">
              Quantity
            </Text>
            <View className="flex-row items-center">
              <TouchableOpacity
                onPress={() => setQuantity(Math.max(1, quantity - 1))}
                disabled={quantity <= 1}
                className={`w-10 h-10 rounded-xl items-center justify-center ${
                  quantity <= 1 ? "bg-slate-50" : "bg-slate-100"
                }`}
              >
                <Ionicons
                  name="remove"
                  size={20}
                  color={quantity <= 1 ? "#cbd5e1" : "#475569"}
                />
              </TouchableOpacity>
              <Text className="mx-5 text-lg font-bold text-slate-900 w-8 text-center">
                {quantity}
              </Text>
              <TouchableOpacity
                onPress={() => setQuantity(Math.min(stock, quantity + 1))}
                disabled={quantity >= stock}
                className={`w-10 h-10 rounded-xl items-center justify-center ${
                  quantity >= stock ? "bg-slate-50" : "bg-slate-100"
                }`}
              >
                <Ionicons
                  name="add"
                  size={20}
                  color={quantity >= stock ? "#cbd5e1" : "#475569"}
                />
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </ScrollView>

      {/* Bottom CTA area */}
      <View className="px-5 pb-5 pt-3 border-t border-slate-100">
        {/* Reserve & Pay button */}
        <TouchableOpacity
          onPress={() => setShowForm(true)}
          disabled={stock === 0}
          className={`py-4 rounded-2xl items-center ${
            stock === 0 ? "bg-slate-200" : "bg-[#9333ea]"
          }`}
          activeOpacity={0.85}
        >
          <Text
            className={`text-base font-bold ${
              stock === 0 ? "text-slate-400" : "text-white"
            }`}
          >
            {stock === 0
              ? "Out of Stock"
              : `Reserve & Pay — MVR ${totalPrice.toFixed(2)}`}
          </Text>
        </TouchableOpacity>

        {/* Divider */}
        <View className="flex-row items-center my-3">
          <View className="flex-1 h-px bg-slate-200" />
          <Text className="mx-3 text-xs text-slate-400 font-medium uppercase">
            or contact the seller
          </Text>
          <View className="flex-1 h-px bg-slate-200" />
        </View>

        {/* WhatsApp button */}
        <TouchableOpacity
          onPress={openWhatsApp}
          className="flex-row items-center justify-center py-3.5 rounded-2xl bg-[#25D366]"
          activeOpacity={0.85}
        >
          <Ionicons name="logo-whatsapp" size={20} color="#ffffff" />
          <Text className="text-sm font-bold text-white ml-2">
            WhatsApp — Message Seller
          </Text>
        </TouchableOpacity>
      </View>

      {/* Delivery form modal (bottom sheet) */}
      <Modal visible={showForm} animationType="slide" transparent>
        <KeyboardAvoidingView
          behavior={Platform.OS === "ios" ? "padding" : "height"}
          className="flex-1 justify-end"
        >
          {/* Backdrop */}
          <TouchableOpacity
            activeOpacity={1}
            onPress={() => setShowForm(false)}
            className="flex-1 bg-black/40"
          />

          {/* Sheet content */}
          <View className="bg-white rounded-t-3xl px-5 pt-5 pb-8">
            {/* Handle bar */}
            <View className="w-10 h-1 rounded-full bg-slate-300 self-center mb-4" />

            <View className="flex-row items-center justify-between mb-5">
              <Text className="text-lg font-bold text-slate-900">
                Delivery Details
              </Text>
              <TouchableOpacity onPress={() => setShowForm(false)}>
                <Ionicons name="close" size={24} color="#64748b" />
              </TouchableOpacity>
            </View>

            {/* Full Name */}
            <Text className="text-sm font-medium text-slate-700 mb-1">
              Full Name
            </Text>
            <TextInput
              className="bg-slate-50 rounded-xl px-4 py-3 text-sm text-slate-800 border border-slate-200 mb-3"
              placeholder="Your full name"
              placeholderTextColor="#94a3b8"
              value={customerName}
              onChangeText={setCustomerName}
            />

            {/* Phone */}
            <Text className="text-sm font-medium text-slate-700 mb-1">
              Phone
            </Text>
            <View className="flex-row items-center mb-3">
              <View className="bg-slate-100 rounded-l-xl px-3 py-3 border border-r-0 border-slate-200">
                <Text className="text-sm text-slate-500 font-medium">
                  +960
                </Text>
              </View>
              <TextInput
                className="flex-1 bg-slate-50 rounded-r-xl px-4 py-3 text-sm text-slate-800 border border-slate-200"
                placeholder="7 digits"
                placeholderTextColor="#94a3b8"
                value={customerPhone}
                onChangeText={(t) =>
                  setCustomerPhone(t.replace(/\D/g, "").slice(0, 7))
                }
                keyboardType="number-pad"
                maxLength={7}
              />
            </View>

            {/* Location picker */}
            <Text className="text-sm font-medium text-slate-700 mb-1">
              Location
            </Text>
            <View className="flex-row gap-2 mb-3">
              {LOCATIONS.map((loc) => (
                <TouchableOpacity
                  key={loc}
                  onPress={() => setDeliveryLocation(loc)}
                  className={`flex-1 py-3 rounded-xl border items-center ${
                    deliveryLocation === loc
                      ? "bg-[#9333ea] border-[#9333ea]"
                      : "bg-white border-slate-200"
                  }`}
                >
                  <Text
                    className={`text-sm font-medium ${
                      deliveryLocation === loc
                        ? "text-white"
                        : "text-slate-600"
                    }`}
                  >
                    {loc}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>

            {/* Delivery Address */}
            <Text className="text-sm font-medium text-slate-700 mb-1">
              Delivery Address
            </Text>
            <TextInput
              className="bg-slate-50 rounded-xl px-4 py-3 text-sm text-slate-800 border border-slate-200 mb-5"
              placeholder="House name, road, block..."
              placeholderTextColor="#94a3b8"
              value={deliveryAddress}
              onChangeText={setDeliveryAddress}
              multiline
              numberOfLines={3}
              textAlignVertical="top"
              style={{ minHeight: 72 }}
            />

            {/* Submit button */}
            <TouchableOpacity
              onPress={handleSubmitOrder}
              disabled={submitting}
              className={`py-4 rounded-2xl items-center ${
                submitting ? "bg-[#a855f7]" : "bg-[#9333ea]"
              }`}
              activeOpacity={0.85}
            >
              <Text className="text-base font-bold text-white">
                {submitting
                  ? "Placing Order..."
                  : `Confirm & Pay — MVR ${totalPrice.toFixed(2)}`}
              </Text>
            </TouchableOpacity>
          </View>
        </KeyboardAvoidingView>
      </Modal>
    </SafeAreaView>
  );
}
