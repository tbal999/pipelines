import React from 'react';

const Step2 = ({ prevStep, nextStep, handleChange, values }) => {
  return (
    <div>
      <h2>Step 2</h2>
      <label>
        Last Name:
        <input
          type="text"
          name="lastName"
          value={values.lastName}
          onChange={handleChange}
        />
      </label>
      <button onClick={prevStep}>Back</button>
      <button onClick={nextStep}>Next</button>
    </div>
  );
};

export default Step2;
