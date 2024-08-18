import React from 'react';

const Step3 = ({ prevStep, submit, handleChange, values }) => {
  return (
    <div>
      <h2>Step 3</h2>
      <label>
        Email:
        <input
          type="email"
          name="email"
          value={values.email}
          onChange={handleChange} // This should be properly defined
        />
      </label>
      <button onClick={prevStep}>Back</button>
      <button onClick={submit}>Submit</button>
    </div>
  );
};

export default Step3;
