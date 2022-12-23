import { useState } from "react";
import axios from "axios";

function App() {
  const [name, setName] = useState("");

  const registrationHandler = async () => {
    const result = await axios.get(
      `http://localhost:1323/register/begin/${name}`
    );
    const data = result.data;
    console.log(data);
  };

  return (
    <div>
      <div>
        name:
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        <br />
        <button onClick={registrationHandler}>Registration</button>
      </div>
    </div>
  );
}

export default App;
