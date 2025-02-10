import { Button } from "@/components/ui/button";
import { useState } from "react";

interface Props {
  defaultNum?: number;
}

export function Counter(props: Props) {
  const [count, setCount] = useState(props.defaultNum || 0);

  return (
    <div>
      <h1>{count}</h1>
      <Button onClick={() => setCount(count + 1)}>Click me</Button>
    </div>
  );
}
