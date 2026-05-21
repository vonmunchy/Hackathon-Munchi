import { useState, useMemo } from "react";
import {
  View,
  Text,
  TextInput,
  FlatList,
  ScrollView,
  TouchableOpacity,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useRouter } from "expo-router";
import { useQuery } from "convex/react";
import { api } from "../../convex/_generated/api";
import ProductCard from "../../components/ProductCard";
import LoadingSpinner from "../../components/LoadingSpinner";
import { Ionicons } from "@expo/vector-icons";

export default function MarketplaceScreen() {
  const router = useRouter();
  const products = useQuery(api.marketplace.getAllActiveProducts);
  const [search, setSearch] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);

  const categories = useMemo(() => {
    if (!products) return [];
    const cats = new Set(
      products.map((p) => p.category).filter((c): c is string => Boolean(c))
    );
    return Array.from(cats);
  }, [products]);

  const filteredProducts = useMemo(() => {
    if (!products) return [];
    let result = products;
    if (search.trim()) {
      const q = search.toLowerCase();
      result = result.filter(
        (p) =>
          p.name.toLowerCase().includes(q) ||
          p.storeName.toLowerCase().includes(q)
      );
    }
    if (selectedCategory) {
      result = result.filter((p) => p.category === selectedCategory);
    }
    return result;
  }, [products, search, selectedCategory]);

  const hasActiveFilters = search || selectedCategory;

  const clearFilters = () => {
    setSearch("");
    setSelectedCategory(null);
  };

  if (products === undefined) {
    return (
      <SafeAreaView className="flex-1 bg-slate-50" edges={["top"]}>
        <LoadingSpinner message="Loading marketplace..." />
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView className="flex-1 bg-slate-50" edges={["top"]}>
      {/* Header */}
      <View className="px-5 pt-4 pb-1">
        <Text className="text-2xl font-bold text-slate-900">Marketplace</Text>
        <Text className="text-sm text-slate-500 mt-0.5">
          Shop from local sellers, pay instantly with Swipe
        </Text>
      </View>

      {/* Search bar */}
      <View className="px-5 mt-3 mb-3">
        <View className="flex-row items-center bg-white rounded-xl border border-slate-200 px-3">
          <Ionicons name="search" size={18} color="#94a3b8" />
          <TextInput
            className="flex-1 py-3 px-2 text-sm text-slate-800"
            placeholder="Search products or stores..."
            placeholderTextColor="#94a3b8"
            value={search}
            onChangeText={setSearch}
            autoCorrect={false}
          />
          {search ? (
            <TouchableOpacity onPress={() => setSearch("")}>
              <Ionicons name="close-circle" size={18} color="#cbd5e1" />
            </TouchableOpacity>
          ) : null}
        </View>
      </View>

      {/* Category filter chips */}
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        className="mb-3"
        contentContainerStyle={{ paddingHorizontal: 20, gap: 8 }}
      >
        <TouchableOpacity
          onPress={() => setSelectedCategory(null)}
          className={`px-4 py-2 rounded-full ${
            !selectedCategory
              ? "bg-[#7c3aed]"
              : "bg-white border border-slate-200"
          }`}
        >
          <Text
            className={`text-sm font-medium ${
              !selectedCategory ? "text-white" : "text-slate-600"
            }`}
          >
            All
          </Text>
        </TouchableOpacity>
        {categories.map((cat) => (
          <TouchableOpacity
            key={cat}
            onPress={() =>
              setSelectedCategory(selectedCategory === cat ? null : cat)
            }
            className={`px-4 py-2 rounded-full ${
              selectedCategory === cat
                ? "bg-[#7c3aed]"
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

      {/* Filter count + clear */}
      {hasActiveFilters ? (
        <View className="px-5 mb-2 flex-row items-center">
          <Text className="text-sm text-slate-500 flex-1">
            {filteredProducts.length} result
            {filteredProducts.length !== 1 ? "s" : ""}
          </Text>
          <TouchableOpacity onPress={clearFilters}>
            <Text className="text-sm font-medium text-[#7c3aed]">
              Clear filters
            </Text>
          </TouchableOpacity>
        </View>
      ) : null}

      {/* Product grid or empty state */}
      {filteredProducts.length === 0 ? (
        <View className="flex-1 items-center justify-center px-8">
          <Ionicons
            name={search ? "search-outline" : "bag-outline"}
            size={48}
            color="#cbd5e1"
          />
          <Text className="text-lg font-semibold text-slate-700 text-center mt-3">
            {search ? "No products match your search" : "No products yet"}
          </Text>
          <Text className="text-sm text-slate-400 text-center mt-1">
            {search
              ? "Try a different search term"
              : "Check back later for new listings"}
          </Text>
        </View>
      ) : (
        <FlatList
          data={filteredProducts}
          numColumns={2}
          keyExtractor={(item) => item._id}
          contentContainerStyle={{ paddingHorizontal: 12, paddingBottom: 20 }}
          columnWrapperStyle={{ gap: 6 }}
          ItemSeparatorComponent={() => <View style={{ height: 6 }} />}
          renderItem={({ item }) => (
            <ProductCard
              name={item.name}
              imageUrl={item.imageUrls[0] ?? ""}
              price={item.basePrice}
              storeName={item.storeName}
              category={item.category}
              onPress={() => router.push(`/product/${item._id}`)}
            />
          )}
        />
      )}
    </SafeAreaView>
  );
}
