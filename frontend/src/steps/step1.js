import React from 'react';

const Step1 = ({ nextStep, handleChange, values }) => {
  return (
    <div>
      <h2>Step 1</h2>
      <label>
        First Name:
        <input
          type="text"
          name="firstName"
          value={values.firstName}
          onChange={handleChange}
        />
      </label>
      <button onClick={nextStep}>Next</button>
    </div>
  );
};

export default Step1;
