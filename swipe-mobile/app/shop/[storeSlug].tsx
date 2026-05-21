import { useState, useMemo } from "react";
import {
  View,
  Text,
  FlatList,
  ScrollView,
  TouchableOpacity,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useRouter, useLocalSearchParams } from "expo-router";
import { useQuery } from "convex/react";
import { api } from "../../convex/_generated/api";
import ProductCard from "../../components/ProductCard";
import LoadingSpinner from "../../components/LoadingSpinner";
import { Ionicons } from "@expo/vector-icons";

export default function StorePage() {
  const router = useRouter();
  const { storeSlug } = useLocalSearchParams<{ storeSlug: string }>();
  const data = useQuery(
    api.products.listByStoreSlugWithVariants,
    storeSlug ? { storeSlug } : "skip"
  );

  const [selectedCategory, setSelectedCategory] = useState("All");

  const categories = useMemo(() => {
    if (!data?.products) return ["All"];
    const cats = new Set(
      data.products
        .map((p) => p.category)
        .filter((c): c is string => !!c)
    );
    return ["All", ...Array.from(cats)];
  }, [data]);

  const filteredProducts = useMemo(() => {
    if (!data?.products) return [];
    if (selectedCategory === "All") return data.products;
    return data.products.filter((p) => p.category === selectedCategory);
  }, [data, selectedCategory]);

  if (data === undefined) {
    return (
      <SafeAreaView className="flex-1 bg-slate-50">
        <LoadingSpinner message="Loading store..." />
      </SafeAreaView>
    );
  }

  if (data === null) {
    return (
      <SafeAreaView className="flex-1 bg-slate-50 items-center justify-center px-8">
        <Text className="text-lg font-semibold text-slate-700">Store not found</Text>
        <TouchableOpacity onPress={() => router.back()} className="mt-4">
          <Text className="text-amethyst-600 font-medium">Go back</Text>
        </TouchableOpacity>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView className="flex-1 bg-slate-50" edges={["top"]}>
      {/* Header */}
      <View className="flex-row items-center px-4 py-3 bg-white border-b border-slate-100">
        <TouchableOpacity onPress={() => router.back()} className="mr-3">
          <Ionicons name="arrow-back" size={24} color="#1e293b" />
        </TouchableOpacity>
        <View className="flex-1">
          <Text className="text-lg font-bold text-slate-900" numberOfLines={1}>
            {data.store.name}
          </Text>
          {data.store.description ? (
            <Text className="text-xs text-slate-400" numberOfLines={1}>
              {data.store.description}
            </Text>
          ) : null}
        </View>
      </View>

      {/* Category filter */}
      {categories.length > 1 ? (
        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          className="px-4 py-3"
          contentContainerStyle={{ gap: 8 }}
        >
          {categories.map((cat) => (
            <TouchableOpacity
              key={cat}
              onPress={() => setSelectedCategory(cat)}
              className={`px-4 py-2 rounded-full ${
                selectedCategory === cat
                  ? "bg-amethyst-600"
                  : "bg-white border border-slate-200"
              }`}
            >
              <Text
                className={`text-sm font-medium ${
                  selectedCategory === cat ? "text-white" : "text-slate-600"
                }`}
              >
                {cat}
              </Text>
            </TouchableOpacity>
          ))}
        </ScrollView>
      ) : null}

      {/* Product grid */}
      {filteredProducts.length === 0 ? (
        <View className="flex-1 items-center justify-center px-8">
          <Text className="text-5xl mb-3">🏪</Text>
          <Text className="text-lg font-semibold text-slate-700 text-center">
            No products yet
          </Text>
          <Text className="text-sm text-slate-400 text-center mt-1">
            This store hasn't listed any products.
          </Text>
        </View>
      ) : (
        <FlatList
          data={filteredProducts}
          numColumns={2}
          keyExtractor={(item) => item._id}
          contentContainerStyle={{ paddingHorizontal: 8, paddingBottom: 20 }}
          renderItem={({ item }) => (
            <ProductCard
              name={item.name}
              imageUrl={item.imageUrls[0] ?? ""}
              price={item.basePrice}
              storeName={data.store.name}
              category={item.category}
              onPress={() => router.push(`/product/${item._id}`)}
            />
          )}
        />
      )}
    </SafeAreaView>
  );
}
