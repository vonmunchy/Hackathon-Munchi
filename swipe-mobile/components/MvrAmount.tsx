import { Text } from "react-native";

interface MvrAmountProps {
  amount: number;
  className?: string;
}

export default function MvrAmount({ amount, className }: MvrAmountProps) {
  return (
    <Text className={className ?? "text-base font-bold text-slate-900"}>
      MVR {amount.toFixed(2)}
    </Text>
  );
}
